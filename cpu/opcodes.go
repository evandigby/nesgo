package cpu

import (
	"errors"
	"fmt"
	"strings"
)

var ErrorUnknownAddressMode = errors.New("Unknown Addressing Mode")
var ErrorUnknownInstruction = errors.New("Unknown Instruction")

type Executer func(s *State) int

type Opcode struct {
	address     int
	opcode      []byte
	addressMode int
	instruction string
	operands    string
	cycles      int
	executer    Executer
}

func (o *Opcode) Address() int        { return o.address }
func (o *Opcode) Operands() string    { return o.operands }
func (o *Opcode) Opcode() []byte      { return o.opcode }
func (o *Opcode) Executer() Executer  { return o.executer }
func (o *Opcode) Instruction() string { return o.instruction }
func (o *Opcode) Disassemble() string {
	return strings.TrimSpace(fmt.Sprintf("%v %v", o.instruction, o.operands))
}

func NewOpcode(memory []byte, address int) *Opcode {
	op := memory[address]

	addressMode := getAddressMode(op)

	opcode := getOpcode(memory, address, addressMode)

	instruction := getInstruction(op)

	cycles := getCycles()

	value := getValue(opcode)

	return &Opcode{
		address,
		opcode,
		addressMode,
		instruction,
		getOperands(addressMode, value),
		getCycles(),
		getExecuter(instruction, addressMode, value, cycles),
	}
}

func getInstruction(op byte) string {
	switch op {
	case 0x69, 0x65, 0x75, 0x6D, 0x7D, 0x79, 0x61, 0x71:
		return opADC
	case 0x29, 0x25, 0x35, 0x2D, 0x3D, 0x39, 0x21, 0x31:
		return opAND
	case 0x0A, 0x06, 0x16, 0x0E, 0x1E:
		return opASL
	case 0x90:
		return opBCC
	case 0xB0:
		return opBCS
	case 0xF0:
		return opBEQ
	case 0x24, 0x2C:
		return opBIT
	case 0x30:
		return opBMI
	case 0xD0:
		return opBNE
	case 0x10:
		return opBPL
	case 0x00:
		return opBRK
	case 0x50:
		return opBVC
	case 0x70:
		return opBVS
	case 0x18:
		return opCLC
	case 0xD8:
		return opCLD
	case 0x58:
		return opCLI
	case 0xB8:
		return opCLV
	case 0xC9, 0xC5, 0xD5, 0xCD, 0xDD, 0xD9, 0xC1, 0xD1:
		return opCMP
	case 0xE0, 0xE4, 0xEC:
		return opCPX
	case 0xC0, 0xC4, 0xCC:
		return opCPY
	case 0xC6, 0xD6, 0xCE, 0xDE:
		return opDEC
	case 0xCA:
		return opDEX
	case 0x88:
		return opDEY
	case 0x49, 0x45, 0x55, 0x4D, 0x5D, 0x59, 0x41, 0x51:
		return opEOR
	case 0xE6, 0xF6, 0xEE, 0xFE:
		return opINC
	case 0xE8:
		return opINX
	case 0xC8:
		return opINY
	case 0x4C, 0x6C:
		return opJMP
	case 0x20:
		return opJSR
	case 0xA9, 0xA5, 0xB5, 0xAD, 0xBD, 0xB9, 0xA1, 0xB1:
		return opLDA
	case 0xA2, 0xA6, 0xB6, 0xAE, 0xBE:
		return opLDX
	case 0xA0, 0xA4, 0xB4, 0xAC, 0xBC:
		return opLDY
	case 0x4A, 0x46, 0x56, 0x4E, 0x5E:
		return opLSR
	case 0xEA:
		return opNOP
	case 0x09, 0x05, 0x15, 0x0D, 0x1D, 0x19, 0x01, 0x11:
		return opORA
	case 0x48:
		return opPHA
	case 0x08:
		return opPHP
	case 0x68:
		return opPLA
	case 0x28:
		return opPLP
	case 0x2A, 0x26, 0x36, 0x2E, 0x3E:
		return opROL
	case 0x6A, 0x66, 0x76, 0x6E, 0x7E:
		return opROR
	case 0x40:
		return opRTI
	case 0x60:
		return opRTS
	case 0xE9, 0xE5, 0xF5, 0xED, 0xFD, 0xF9, 0xE1, 0xF1:
		return opSBC
	case 0x38:
		return opSEC
	case 0xF8:
		return opSED
	case 0x78:
		return opSEI
	case 0x85, 0x95, 0x8D, 0x9D, 0x99, 0x81, 0x91:
		return opSTA
	case 0x86, 0x96, 0x8E:
		return opSTX
	case 0x84, 0x94, 0x8C:
		return opSTY
	case 0xAA:
		return opTAX
	case 0xA8:
		return opTAY
	case 0xBA:
		return opTSX
	case 0x8A:
		return opTXA
	case 0x9A:
		return opTXS
	case 0x98:
		return opTYA
	default:
		return opUNK
	}
}

func getValue(opcode []byte) uint16 {
	if len(opcode) > 2 {
		return (uint16(opcode[2]) << 8) | uint16(opcode[1])
	} else if len(opcode) > 1 {
		return uint16(opcode[1])
	} else {
		return uint16(0)
	}
}
func getCycles() int {
	return 2
}

func getOperands(addressMode int, operand uint16) string {
	switch addressMode {
	case AddressImplied:
		return ""
	case AddressAccumulator:
		return "A"
	case AddressRelative, AddressZeroPage, AddressAbsolute:
		return fmt.Sprintf("$%X", operand)
	case AddressZeroPageX, AddressAbsoluteX:
		return fmt.Sprintf("$%X, X", operand)
	case AddressZeroPageY, AddressAbsoluteY:
		return fmt.Sprintf("$%X, Y", operand)
	case AddressImmediate:
		return fmt.Sprintf("#$%X", operand)
	case AddressIndirectX:
		return fmt.Sprintf("($%X, X)", operand)
	case AddressIndirectY:
		return fmt.Sprintf("($%X, Y)", operand)
	case AddressIndirect:
		return fmt.Sprintf("($%X)", operand)
	}

	// Should never get here
	panic(ErrorUnknownAddressMode)
}

func getAddressMode(op byte) int {
	// Exceptions to the rules
	switch op {
	case 0x00, 0x60:
		return AddressImplied
	case 0x20, 0x40, 0x80:
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

func getOpcode(memory []byte, address int, addressMode int) []byte {

	switch addressMode {
	case AddressImplied,
		AddressAccumulator:
		return memory[address : address+1]

	case AddressRelative,
		AddressImmediate,
		AddressZeroPage,
		AddressZeroPageX,
		AddressZeroPageY,
		AddressIndirectX,
		AddressIndirectY:

		return memory[address : address+2]

	case AddressAbsolute,
		AddressAbsoluteX,
		AddressAbsoluteY,
		AddressIndirect:

		return memory[address : address+3]

	}

	panic(ErrorUnknownAddressMode)
}

func getExecuter(instruction string, addressMode int, value uint16, cycles int) Executer {
	switch instruction {
	case opADC:
		return ADC(addressMode, value, cycles)
	case opAND:
		return AND(addressMode, value, cycles)
	case opASL:
		return ASL(addressMode, value, cycles)
	case opBCC:
		return BCC(addressMode, value, cycles)
	case opBCS:
		return BCS(addressMode, value, cycles)
	case opBEQ:
		return BEQ(addressMode, value, cycles)
	case opBIT:
		return BIT(addressMode, value, cycles)
	case opBMI:
		return BMI(addressMode, value, cycles)
	case opBNE:
		return BNE(addressMode, value, cycles)
	case opBPL:
		return BPL(addressMode, value, cycles)
	case opBRK:
		return BRK(addressMode, value, cycles)
	case opBVC:
		return BVC(addressMode, value, cycles)
	case opBVS:
		return BVS(addressMode, value, cycles)
	case opCLC:
		return CLC(addressMode, value, cycles)
	case opCLD:
		return CLD(addressMode, value, cycles)
	case opCLI:
		return CLI(addressMode, value, cycles)
	case opCLV:
		return CLV(addressMode, value, cycles)
	case opCMP:
		return CMP(addressMode, value, cycles)
	case opCPX:
		return CPX(addressMode, value, cycles)
	case opCPY:
		return CPY(addressMode, value, cycles)
	case opDEC:
		return DEC(addressMode, value, cycles)
	case opDEX:
		return DEX(addressMode, value, cycles)
	case opDEY:
		return DEY(addressMode, value, cycles)
	case opEOR:
		return EOR(addressMode, value, cycles)
	case opINC:
		return INC(addressMode, value, cycles)
	case opINX:
		return INX(addressMode, value, cycles)
	case opINY:
		return INY(addressMode, value, cycles)
	case opJMP:
		return JMP(addressMode, value, cycles)
	case opJSR:
		return JSR(addressMode, value, cycles)
	case opLDA:
		return LDA(addressMode, value, cycles)
	case opLDX:
		return LDX(addressMode, value, cycles)
	case opLDY:
		return LDY(addressMode, value, cycles)
	case opLSR:
		return LSR(addressMode, value, cycles)
	case opNOP:
		return NOP(addressMode, value, cycles)
	case opORA:
		return ORA(addressMode, value, cycles)
	case opPHA:
		return PHA(addressMode, value, cycles)
	case opPHP:
		return PHP(addressMode, value, cycles)
	case opPLA:
		return PLA(addressMode, value, cycles)
	case opPLP:
		return PLP(addressMode, value, cycles)
	case opROL:
		return ROL(addressMode, value, cycles)
	case opROR:
		return ROR(addressMode, value, cycles)
	case opRTI:
		return RTI(addressMode, value, cycles)
	case opRTS:
		return RTS(addressMode, value, cycles)
	case opSBC:
		return SBC(addressMode, value, cycles)
	case opSEC:
		return SEC(addressMode, value, cycles)
	case opSED:
		return SED(addressMode, value, cycles)
	case opSEI:
		return SEI(addressMode, value, cycles)
	case opSTA:
		return STA(addressMode, value, cycles)
	case opSTX:
		return STX(addressMode, value, cycles)
	case opSTY:
		return STY(addressMode, value, cycles)
	case opTAX:
		return TAX(addressMode, value, cycles)
	case opTAY:
		return TAY(addressMode, value, cycles)
	case opTSX:
		return TSX(addressMode, value, cycles)
	case opTXA:
		return TXA(addressMode, value, cycles)
	case opTXS:
		return TXS(addressMode, value, cycles)
	case opTYA:
		return TYA(addressMode, value, cycles)
	default:
		return NOP(addressMode, value, cycles)
	}
}

// Decompile will Read each address in memory as though it could be executed.
// This will produce many indvalid opcodes, but doesn't require
// us to differentiate between instructions and data ahead of time.
// Any invalid opcodes will result in a "nop". Hopefully the code never jumps to them :)
// Perhaps we should put a panicking instruction to suss these out.
func Decompile(memory []byte) []*Opcode {
	codes := make([]*Opcode, len(memory))

	for i := 0; i < len(memory); i++ {
		codes[i] = NewOpcode(memory, i)
	}

	return codes
}

func Execution(opcodes []Opcode) []Executer {
	e := make([]Executer, len(opcodes))
	for i, o := range opcodes {
		e[i] = o.Executer()
	}
	return e
}
