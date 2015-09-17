package cpu

import (
	"errors"
	"fmt"
	"strings"
)

var ErrorUnknownAddressMode = errors.New("Unknown Addressing Mode")
var ErrorUnknownInstruction = errors.New("Unknown Instruction")

type Executer func(s *State) (cycles int, nexPc uint16)

type Opcode struct {
	address     uint16
	opcode      []*byte
	addressMode int
	instruction string
	cycles      int
	executer    Executer
	getter      Getter
	setter      Setter
}

func (o *Opcode) Address() uint16  { return o.address }
func (o *Opcode) AddressMode() int { return o.addressMode }
func (o *Opcode) Operand() uint16  { return getValue(o.opcode) }
func (o *Opcode) Operands() string {
	value := getValue(o.opcode)
	operands := getOperands(o.addressMode, o.address, uint16(len(o.opcode)), value)
	return operands
}
func (o *Opcode) Opcode() []*byte     { return o.opcode }
func (o *Opcode) Executer() Executer  { return o.executer }
func (o *Opcode) Instruction() string { return o.instruction }
func (o *Opcode) Cycles() int         { return o.cycles }
func (o *Opcode) Get(s *State) byte {
	v, _ := o.getter(s)
	return v
}

func (o *Opcode) Bytes() string {
	op := ""
	for _, v := range o.opcode {
		op += fmt.Sprintf("%02X ", *v)
	}

	return strings.TrimSpace(op)
}
func (o *Opcode) Disassemble() string {
	return strings.TrimSpace(fmt.Sprintf("%v %v", o.instruction, o.Operands()))
}

func NewOpcode(memory []*byte, address uint16) *Opcode {
	op := *memory[address]

	addressMode := getAddressMode(op)

	opcode := getOpcode(memory, address, addressMode)
	l := uint16(len(opcode))
	if l == 0 {
		return nil
	}
	instruction := getInstruction(op)
	cycles := getCycles(instruction, addressMode)
	value := getValue(opcode)

	getter := getGetter(addressMode, address, l, value)
	setter := getSetter(addressMode, address, l, value)

	return &Opcode{
		address,
		opcode,
		addressMode,
		instruction,
		cycles,
		getExecuter(instruction, opcode, addressMode, address, value, cycles, getter, setter),
		getter,
		setter,
	}
}

// Decompile will Read each address in memory as though it could be executed.
// This will produce many indvalid opcodes, but doesn't require
// us to differentiate between instructions and data ahead of time.
// Any invalid opcodes will result in a "nop". Hopefully the code never jumps to them :)
// Perhaps we should put a panicking instruction to suss these out.
func Decompile(memory []*byte) []*Opcode {
	codes := make([]*Opcode, len(memory))

	for i := 0; i < len(memory); i++ {
		codes[i] = NewOpcode(memory, uint16(i))
	}

	return codes
}

func Execution(opcodes []*Opcode) []Executer {
	e := make([]Executer, len(opcodes))
	for i, o := range opcodes {
		e[i] = o.Executer()
	}
	return e
}

func getValue(opcode []*byte) uint16 {
	if len(opcode) > 2 {
		return (uint16(*opcode[2]) << 8) | uint16(*opcode[1])
	} else if len(opcode) > 1 {
		return uint16(*opcode[1])
	} else {
		return uint16(0)
	}
}

func getOperands(addressMode int, address, instructionLength, operand uint16) string {
	f := "%02X"
	if instructionLength == 3 {
		f = "%04X"
	}
	switch addressMode {
	case AddressImplied:
		return ""
	case AddressAccumulator:
		return "A"
	case AddressRelative:
		a, _ := calculateRelativeAddress(instructionLength, operand, uint16(address))
		return fmt.Sprintf(fmt.Sprintf("$%s", f), a)
	case AddressZeroPage, AddressAbsolute, AddressAddress:
		return fmt.Sprintf(fmt.Sprintf("$%s", f), operand)
	case AddressZeroPageX, AddressAbsoluteX:
		return fmt.Sprintf(fmt.Sprintf("$%s, X", f), operand)
	case AddressZeroPageY, AddressAbsoluteY:
		return fmt.Sprintf(fmt.Sprintf("$%s, Y", f), operand)
	case AddressImmediate:
		return fmt.Sprintf(fmt.Sprintf("#$%s", f), operand)
	case AddressIndirectX:
		return fmt.Sprintf(fmt.Sprintf("($%s, X)", f), operand)
	case AddressIndirectY:
		return fmt.Sprintf(fmt.Sprintf("($%s, Y)", f), operand)
	case AddressIndirect:
		return fmt.Sprintf(fmt.Sprintf("($%s)", f), operand)
	}

	// Should never get here
	panic(ErrorUnknownAddressMode)
}

func getAddressMode(op byte) int {
	// Exceptions to the rules
	switch op {
	case 0x20, 0x4C:
		return AddressAddress
	case 0x00, 0x60:
		return AddressImplied
	case 0x40, 0x80:
		return AddressAbsolute
	case 0xA0, 0xC0, 0xE0:
		return AddressImmediate
	case 0x96, 0x97, 0x9E, 0x9F, 0xB6, 0xB7, 0xBE, 0xBF:
		return AddressZeroPageY
	case 0x6C:
		return AddressIndirect
	case 0x0A, 0x2A, 0x4A, 0x6A:
		return AddressAccumulator
	}

	var off uint16

	opint16 := uint16(op)
	for i := uint16(0x00); i <= 0xE0; i = i + 0x20 {
		if opint16 < (i + 0x20) {
			off = opint16 - i
			break
		}
	}

	switch off {
	case 0x01, 0x03:
		return AddressIndirectX
	case 0x08, 0x0A, 0x18, 0x1A:
		return AddressImplied
	case 0x04, 0x05, 0x06, 0x07:
		return AddressZeroPage
	case 0x02, 0x09, 0x0B:
		return AddressImmediate
	case 0x0C, 0x0D, 0x0E, 0x0F:
		return AddressAbsolute
	case 0x10:
		return AddressRelative
	case 0x11, 0x13:
		return AddressIndirectY
	case 0x14, 0x15, 0x16, 0x17:
		return AddressZeroPageX
	case 0x19, 0x1B:
		return AddressAbsoluteY
	case 0x1C, 0x1D, 0x1E, 0x1F:
		return AddressAbsoluteX
	default:
		return AddressImplied // Default to a single byte instruction
	}
}

func getOpcode(memory []*byte, address uint16, addressMode int) []*byte {

	var offset int
	switch addressMode {
	case AddressImplied,
		AddressAccumulator:
		offset = 1

	case AddressRelative,
		AddressImmediate,
		AddressZeroPage,
		AddressZeroPageX,
		AddressZeroPageY,
		AddressIndirectX,
		AddressIndirectY:

		offset = 2

	case AddressAbsolute,
		AddressAddress,
		AddressAbsoluteX,
		AddressAbsoluteY,
		AddressIndirect:

		offset = 3
	}

	start := int(address)
	end := start + offset
	if end < len(memory) {
		return memory[start:end]
	}

	return []*byte{}
}

func getExecuter(instruction string, opcode []*byte, addressMode int, address, value uint16, cycles int, getter Getter, setter Setter) Executer {
	length := uint16(len(opcode))
	switch instruction {
	case opADC:
		return ADC(getter, setter, address, length, value, cycles)
	case opAND:
		return AND(getter, setter, address, length, value, cycles)
	case opASL:
		return ASL(getter, setter, address, length, value, cycles)
	case opBCC:
		return BCC(getter, setter, address, length, value, cycles)
	case opBCS:
		return BCS(getter, setter, address, length, value, cycles)
	case opBEQ:
		return BEQ(getter, setter, address, length, value, cycles)
	case opBIT:
		return BIT(getter, setter, address, length, value, cycles)
	case opBMI:
		return BMI(getter, setter, address, length, value, cycles)
	case opBNE:
		return BNE(getter, setter, address, length, value, cycles)
	case opBPL:
		return BPL(getter, setter, address, length, value, cycles)
	case opBRK:
		return BRK(getter, setter, address, length, value, cycles)
	case opBVC:
		return BVC(getter, setter, address, length, value, cycles)
	case opBVS:
		return BVS(getter, setter, address, length, value, cycles)
	case opCLC:
		return CLC(getter, setter, address, length, value, cycles)
	case opCLD:
		return CLD(getter, setter, address, length, value, cycles)
	case opCLI:
		return CLI(getter, setter, address, length, value, cycles)
	case opCLV:
		return CLV(getter, setter, address, length, value, cycles)
	case opCMP:
		return CMP(getter, setter, address, length, value, cycles)
	case opCPX:
		return CPX(getter, setter, address, length, value, cycles)
	case opCPY:
		return CPY(getter, setter, address, length, value, cycles)
	case opDEC:
		return DEC(getter, setter, address, length, value, cycles)
	case opDEX:
		return DEX(getter, setter, address, length, value, cycles)
	case opDEY:
		return DEY(getter, setter, address, length, value, cycles)
	case opEOR:
		return EOR(getter, setter, address, length, value, cycles)
	case opINC:
		return INC(getter, setter, address, length, value, cycles)
	case opINX:
		return INX(getter, setter, address, length, value, cycles)
	case opINY:
		return INY(getter, setter, address, length, value, cycles)
	case opJMP:
		return JMP(getter, setter, address, length, value, cycles)
	case opJSR:
		return JSR(getter, setter, address, length, value, cycles)
	case opLDA:
		return LDA(getter, setter, address, length, value, cycles)
	case opLDX:
		return LDX(getter, setter, address, length, value, cycles)
	case opLDY:
		return LDY(getter, setter, address, length, value, cycles)
	case opLSR:
		return LSR(getter, setter, address, length, value, cycles)
	case opNOP:
		return NOP(getter, setter, address, length, value, cycles)
	case opORA:
		return ORA(getter, setter, address, length, value, cycles)
	case opPHA:
		return PHA(getter, setter, address, length, value, cycles)
	case opPHP:
		return PHP(getter, setter, address, length, value, cycles)
	case opPLA:
		return PLA(getter, setter, address, length, value, cycles)
	case opPLP:
		return PLP(getter, setter, address, length, value, cycles)
	case opROL:
		return ROL(getter, setter, address, length, value, cycles)
	case opROR:
		return ROR(getter, setter, address, length, value, cycles)
	case opRTI:
		return RTI(getter, setter, address, length, value, cycles)
	case opRTS:
		return RTS(getter, setter, address, length, value, cycles)
	case opSBC:
		return SBC(getter, setter, address, length, value, cycles)
	case opSEC:
		return SEC(getter, setter, address, length, value, cycles)
	case opSED:
		return SED(getter, setter, address, length, value, cycles)
	case opSEI:
		return SEI(getter, setter, address, length, value, cycles)
	case opSTA:
		return STA(getter, setter, address, length, value, cycles)
	case opSTX:
		return STX(getter, setter, address, length, value, cycles)
	case opSTY:
		return STY(getter, setter, address, length, value, cycles)
	case opTAX:
		return TAX(getter, setter, address, length, value, cycles)
	case opTAY:
		return TAY(getter, setter, address, length, value, cycles)
	case opTSX:
		return TSX(getter, setter, address, length, value, cycles)
	case opTXA:
		return TXA(getter, setter, address, length, value, cycles)
	case opTXS:
		return TXS(getter, setter, address, length, value, cycles)
	case opTYA:
		return TYA(getter, setter, address, length, value, cycles)
	default:
		return NOP(getter, setter, address, length, value, cycles)
	}
}
