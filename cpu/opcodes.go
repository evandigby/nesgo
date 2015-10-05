package cpu

import (
	"errors"
	"fmt"
	"strings"
)

var ErrorUnknownAddressMode = errors.New("Unknown Addressing Mode")
var ErrorUnknownInstruction = errors.New("Unknown Instruction")

type Executer func(c *CPU) (cycles int, nexPc uint16)

type Opcode struct {
	address             uint16
	opcode              []*byte
	addressMode         int
	instruction         string
	cycles              int
	length              uint16
	value               uint16
	getter              Getter
	setter              Setter
	addressGetter       AddressGetter
	addressGetterWrong  AddressGetter
	intermediateAddress AddressGetter
}

func (o *Opcode) Address() uint16  { return o.address }
func (o *Opcode) AddressMode() int { return o.addressMode }
func (o *Opcode) Operand() uint16  { return getValue(o.opcode) }
func (o *Opcode) Operands() string {
	operands := getOperands(o.addressMode, o.address, uint16(len(o.opcode)), o.value)
	return operands
}
func (o *Opcode) Opcode() []*byte { return o.opcode }

func (o *Opcode) Instruction() string { return o.instruction }
func (o *Opcode) Cycles() int         { return o.cycles }
func (o *Opcode) Get(c *CPU) byte {
	v, _ := o.getter(c)
	return v
}
func (o *Opcode) AddressGet(c *CPU) (uint16, bool) {
	return o.addressGetter(c)
}

func (o *Opcode) Bytes() string {
	op := ""
	for _, v := range o.opcode {
		op += fmt.Sprintf("%02X ", *v)
	}

	return strings.TrimSpace(op)
}
func (o *Opcode) Disassemble() string {
	return fmt.Sprintf("%v %v", o.instruction, o.Operands())
}
func (o *Opcode) GetValueAt(c *CPU) string {
	addr, _ := o.addressGetter(c)
	intAddr, _ := o.intermediateAddress(c)
	addrWrong, _ := o.addressGetterWrong(c)
	switch o.addressMode {
	case AddressZeroPage, AddressAbsolute:
		return fmt.Sprintf("= %02X", o.Get(c))
	case AddressZeroPageX, AddressZeroPageY:
		return fmt.Sprintf("@ %02X = %02X", addr, o.Get(c))
	case AddressIndirectX:
		return fmt.Sprintf("@ %02X = %04X = %02X", intAddr, addr, o.Get(c))
	case AddressIndirectY:
		return fmt.Sprintf("= %04X @ %04X = %02X", intAddr, addr, o.Get(c))
	case AddressAbsoluteX, AddressAbsoluteY:
		return fmt.Sprintf("@ %04X = %02X", addr, o.Get(c))
	case AddressIndirect:
		return fmt.Sprintf("= %04X", addrWrong)
	default:
		return ""
	}
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
	operand := getValue(opcode)
	length := uint16(len(opcode))

	getter := getGetter(addressMode, address, l, operand)
	setter := getSetter(addressMode, address, l, operand)
	addressGetter := addressCalculator(addressMode, address, l, operand)
	iaddressGetter := intermediateAddressCalculator(addressMode, l, operand)

	addressGetterWrong := addressGetter
	if addressMode == AddressIndirect {
		addressGetterWrong = addressCalculator(AddressIndirectWrong, address, l, operand)
	}

	return &Opcode{
		address,
		opcode,
		addressMode,
		instruction,
		cycles,
		length,
		operand,
		getter,
		setter,
		addressGetter,
		addressGetterWrong,
		iaddressGetter,
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
		return fmt.Sprintf(fmt.Sprintf("$%s,X", f), operand)
	case AddressZeroPageY, AddressAbsoluteY:
		return fmt.Sprintf(fmt.Sprintf("$%s,Y", f), operand)
	case AddressImmediate:
		return fmt.Sprintf(fmt.Sprintf("#$%s", f), operand)
	case AddressIndirectX:
		return fmt.Sprintf(fmt.Sprintf("($%s,X)", f), operand)
	case AddressIndirectY:
		return fmt.Sprintf(fmt.Sprintf("($%s),Y", f), operand)
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
	case 0x00, 0x60, 0x40:
		return AddressImplied
		//	case 0x80:
		//		return AddressAbsolute
	case 0xA0, 0xC0, 0xE0, 0x80:
		return AddressImmediate
	case 0x96, 0x97, 0x9E, 0x9F, 0xB6, 0xB7:
		return AddressZeroPageY
	case 0xBE, 0xBF:
		return AddressAbsoluteY
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

func (o *Opcode) Executer() Executer {

	switch o.instruction {
	case opADC:
		return ADC(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opAND:
		return AND(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opASL:
		return ASL(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opBCC:
		return BCC(o.address, o.length, o.value, o.cycles)
	case opBCS:
		return BCS(o.address, o.length, o.value, o.cycles)
	case opBEQ:
		return BEQ(o.address, o.length, o.value, o.cycles)
	case opBIT:
		return BIT(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opBMI:
		return BMI(o.address, o.length, o.value, o.cycles)
	case opBNE:
		return BNE(o.address, o.length, o.value, o.cycles)
	case opBPL:
		return BPL(o.address, o.length, o.value, o.cycles)
	case opBRK:
		return BRK(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opBVC:
		return BVC(o.address, o.length, o.value, o.cycles)
	case opBVS:
		return BVS(o.address, o.length, o.value, o.cycles)
	case opCLC:
		return CLC(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opCLD:
		return CLD(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opCLI:
		return CLI(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opCLV:
		return CLV(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opCMP:
		return CMP(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opCPX:
		return CPX(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opCPY:
		return CPY(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opDEC:
		return DEC(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opDEX:
		return DEX(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opDEY:
		return DEY(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opEOR:
		return EOR(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opINC:
		return INC(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opINX:
		return INX(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opINY:
		return INY(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opJMP:
		return JMP(o.addressGetter, o.address, o.length, o.value, o.cycles)
	case opJSR:
		return JSR(o.addressGetter, o.address, o.length, o.value, o.cycles)
	case opLDA:
		return LDA(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opLDX:
		return LDX(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opLDY:
		return LDY(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opLSR:
		return LSR(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opNOP:
		return NOP(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opNOPu:
		return NOPGET(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opORA:
		return ORA(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opPHA:
		return PHA(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opPHP:
		return PHP(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opPLA:
		return PLA(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opPLP:
		return PLP(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opROL:
		return ROL(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opROR:
		return ROR(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opRTI:
		return RTI(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opRTS:
		return RTS(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opSBC, opSBCu:
		return SBC(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opSEC:
		return SEC(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opSED:
		return SED(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opSEI:
		return SEI(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opSTA:
		return STA(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opSTX:
		return STX(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opSTY:
		return STY(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opTAX:
		return TAX(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opTAY:
		return TAY(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opTSX:
		return TSX(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opTXA:
		return TXA(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opTXS:
		return TXS(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opTYA:
		return TYA(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opLAXu:
		return LAX(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opSAXu:
		return SAX(o.getter, o.setter, o.address, o.length, o.value, o.cycles)

	case opAACu:
		return NOP(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opAARu:
		return NOP(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opASRu:
		return NOP(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opATXu:
		return NOP(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opAXSu:
		return NOP(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opDCPu:
		return DCP(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opDOPu:
		return NOP(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opISBu:
		return ISB(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opKILu:
		return NOP(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opLARu:
		return NOP(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opRLAu:
		return RLA(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opRRAu:
		return RRA(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opSLOu:
		return SLO(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opSREu:
		return SRE(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opSXAu:
		return NOP(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opSYAu:
		return NOP(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opTOPu:
		return NOP(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opXAAu:
		return NOP(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	case opXASu:
		return NOP(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	default:
		return NOP(o.getter, o.setter, o.address, o.length, o.value, o.cycles)
	}
}
