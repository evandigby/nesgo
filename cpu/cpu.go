package cpu

import (
	"fmt"
	"strconv"

	"github.com/evandigby/nesgo/rom"
)

const (
	AddressImmediate int = iota
	AddressZeroPage
	AddressZeroPageX
	AddressZeroPageY
	AddressAbsolute
	AddressAbsoluteX
	AddressAbsoluteY
	AddressImplied
	AddressAccumulator
	AddressIndirect
	AddressIndirectX
	AddressIndirectY
	AddressRelative
	AddressAddress
)

type CPU struct {
	State *State
	rom   rom.ROM

	nintendulatorLog bool
	Sync             chan int
}

func NewCPU(r rom.ROM) *CPU {
	return &CPU{NewState(), r, true, make(chan int)}
}

func (c *CPU) loadRom(r rom.ROM) {
	mirror := r.Pages() == 1
	for i, v := range r.ProgramRom() {
		*c.State.Cartridge[i] = *v
		if mirror {
			*c.State.Cartridge[i+0x4000] = *v
		}
	}

	c.State.Opcodes = make([]*Opcode, len(c.State.Memory))
	c.State.Executers = make([]Executer, len(c.State.Memory))
	for i, _ := range c.State.Memory {
		op := NewOpcode(c.State.Memory, uint16(i))
		if op != nil {
			c.State.Opcodes[i] = op
			c.State.Executers[i] = op.Executer()
		}
	}
}

func (c *CPU) getValueAt(addressMode int, instructionLength, operand uint16) string {

	intermediateAddress := c.State.CalculateIntermediateAddress(addressMode, instructionLength, operand)
	address, _ := c.State.CalculateAddress(addressMode, instructionLength, operand)
	value, _ := c.State.GetValue(addressMode, instructionLength, operand)

	switch addressMode {
	case AddressZeroPage:
		return fmt.Sprintf("= %02X", value)
	case AddressZeroPageX, AddressZeroPageY:
		return fmt.Sprintf("@ %02X = %02X", address, value)
	case AddressIndirectX, AddressIndirectY:
		return fmt.Sprintf("@ %02X = %04X = %02X", intermediateAddress, address, value)
	case AddressAbsolute, AddressAbsoluteX, AddressAbsoluteY:
		return fmt.Sprintf("@ %04X = %02X", address, value)
	case AddressIndirect:
		return fmt.Sprintf("= %04X", address)
	default:
		return ""
	}
}

func (c *CPU) execute() {
	c.loadRom(c.rom)
	c.State.PowerUp()
	c.State.PC = 0xC000

	cs := 0
	for {
		<-c.Sync
		if c.nintendulatorLog {
			op := c.State.Opcodes[c.State.PC]
			disassembly := fmt.Sprintf("%v %v", op.Disassemble(), c.getValueAt(op.AddressMode(), uint16(len(op.Opcode())), op.Operand()))
			ppuc := (cs * 3) % 341
			fmt.Printf("%04X  %-9s %-31s A:%02X X:%02X Y:%02X P:%02X SP:%02X CYC:%3s\n", c.State.PC, op.Bytes(), disassembly, c.State.A, c.State.X, c.State.Y, c.State.Status(), c.State.SP, strconv.FormatInt(int64(ppuc), 10))

		}
		cycles, pc := c.State.Executers[c.State.PC](c.State)
		cs += cycles
		c.State.PC = pc

		c.Sync <- cycles
	}
}

func (c *CPU) Run() {
	go c.execute()
}

type Flags struct {
	Carry     bool
	Zero      bool
	Interrupt bool
	Decimal   bool
	Break     bool
	Overflow  bool
	Sign      bool
}

func (f *Flags) Status() byte {
	v := byte(32) // Bit 6 always set
	if f.Carry {
		v |= 1
	}
	if f.Zero {
		v |= 2
	}
	if f.Interrupt {
		v |= 4
	}
	if f.Decimal {
		v |= 8
	}
	if f.Break {
		v |= 16
	}
	if f.Overflow {
		v |= 64
	}
	if f.Sign {
		v |= 128
	}

	return v
}

func (f *Flags) SetStatus(v byte) {
	f.Carry = v&1 != 0
	f.Zero = v&2 != 0
	f.Interrupt = v&4 != 0
	f.Decimal = v&8 != 0
	f.Break = v&16 != 0
	f.Overflow = v&64 != 0
	f.Sign = v&128 != 0
}

func (f *Flags) SetSign(value byte) {
	f.Sign = value&0x80 != 0
}

func (f *Flags) SetZero(value byte) {
	f.Zero = value == 0
}

type Registers struct {
	PC uint16
	A  byte
	X  byte
	Y  byte
	SP byte
}

type State struct {
	Registers
	Flags

	Memory         []*byte    `json:"-"`
	Stack          []*byte    `json:"-"`
	PPURegisters   []*byte    `json:"-"`
	APUIORegisters []*byte    `json:"-"`
	Cartridge      []*byte    `json:"-"`
	Executers      []Executer `json:"-"`
	Opcodes        []*Opcode  `json:"-"`
}

func NewState() *State {
	tm := make([]byte, 0x10000)
	m := make([]*byte, 0x10000)
	for i := range m {
		m[i] = &tm[i]
	}
	// Make mirrored memory
	for i := 1; i <= 3; i++ {
		o := i * 0x0800
		for x := 0; x < 0x0800; x++ {
			m[o+x] = m[x]
		}
	}
	// Make stack helper
	s := m[0x0100:0x200] //m[0x0100:0x01FF]

	// PPU Register helper
	ppu := m[0x2000:0x2007]

	// Make mirrored ppu registers
	for i := 1; i <= 0x1FF8/8; i++ {
		o := 0x2000 + (i * 8)
		for x := range ppu {
			m[o+x] = ppu[x]
		}
	}

	// APU Register helper
	apu := m[0x4000:0x4020]

	// Cartridge Memory helper
	c := m[0x8000:0x10000]
	return &State{Registers{}, Flags{}, m, s, ppu, apu, c, nil, nil}
}

func (s *State) Push(val byte) {
	*s.Stack[s.SP] = val
	s.SP++
}

func (s *State) Pop() byte {
	s.SP--
	return *s.Stack[s.SP]
}

func (s *State) PowerUp() {
	s.PC = 0x34
	s.A = 0
	s.X = 0
	s.Y = 0
	s.SP = 0xFD
	s.Interrupt = true
}

func (s *State) Reset() {
	s.SP -= 3
	s.Interrupt = true
}

func calculateRelativeAddress(instructionLength, offset, pc uint16) (uint16, bool) {
	if offset&0x80 == 0 {
		return pc + instructionLength + (offset & 0x7F), false
	} else {
		return pc - (offset & 0x7F), false
	}
}

func (s *State) CalculateIntermediateAddress(addressMode int, instructionLength, offset uint16) uint16 {
	switch addressMode {
	case AddressIndirect:
		return offset
	case AddressIndirectX:
		return uint16(byte(offset) + s.X)
	case AddressIndirectY:
		lsb := *s.Memory[byte(offset)]
		msb := *s.Memory[byte(offset)+byte(1)]
		return (uint16(msb) << 8) | uint16(*s.Memory[lsb])
	}

	return 0
}

func (s *State) CalculateAddress(addressMode int, instructionLength, offset uint16) (address uint16, pageCrossed bool) {
	switch addressMode {
	case AddressZeroPage:
		return uint16(byte(offset)), false
	case AddressAbsolute, AddressAddress:
		return offset, false
	case AddressZeroPageX:
		return uint16(byte(offset) + s.X), false
	case AddressZeroPageY:
		return uint16(byte(offset) + s.Y), false
	case AddressAbsoluteX:
		return offset + uint16(s.X), false
	case AddressAbsoluteY:
		return offset + uint16(s.Y), false
	case AddressIndirect:
		return (uint16(*s.Memory[offset+1]) << 8) | uint16(*s.Memory[offset]), false
	case AddressIndirectX:
		lsb := byte(offset) + s.X
		msb := lsb + 1
		return (uint16(*s.Memory[msb]) << 8) | uint16(*s.Memory[lsb]), false
	case AddressIndirectY:
		lsb := *s.Memory[byte(offset)]
		msb := *s.Memory[byte(offset)+1]
		return ((uint16(msb) << 8) | uint16(*s.Memory[lsb])) + uint16(s.Y), false
	case AddressRelative:
		return calculateRelativeAddress(instructionLength, offset, s.PC)
	}

	return 0, false
}

type Getter func(s *State) (value byte, pageCrossed bool)
type Setter func(s *State, value byte) bool

func getGetter(addressMode int, address, instructionLength, operand uint16) Getter {
	switch addressMode {
	case AddressImmediate:
		return func(s *State) (value byte, pageCrossed bool) { return byte(operand), false }
	case AddressAccumulator:
		return func(s *State) (value byte, pageCrossed bool) { return s.A, false }
	case AddressImplied:
		return func(s *State) (value byte, pageCrossed bool) { return 0, false } // Not sure why we should ever get here
	case AddressZeroPage:
		addr := byte(operand)
		return func(s *State) (value byte, pageCrossed bool) { return *s.Memory[addr], false }
	case AddressAbsolute, AddressAddress:
		return func(s *State) (value byte, pageCrossed bool) { return *s.Memory[operand], false }
	case AddressZeroPageX:
		return func(s *State) (value byte, pageCrossed bool) { return *s.Memory[byte(operand)+s.X], false }
	case AddressZeroPageY:
		return func(s *State) (value byte, pageCrossed bool) { return *s.Memory[byte(operand)+s.Y], false }
	case AddressAbsoluteX:
		return func(s *State) (value byte, pageCrossed bool) { return *s.Memory[operand+uint16(s.X)], false }
	case AddressAbsoluteY:
		return func(s *State) (value byte, pageCrossed bool) { return *s.Memory[operand+uint16(s.Y)], false }
	case AddressIndirect:
		return func(s *State) (value byte, pageCrossed bool) {
			return *s.Memory[(uint16(*s.Memory[operand+1])<<8)|uint16(*s.Memory[operand])], false
		}
	case AddressIndirectX:
		return func(s *State) (value byte, pageCrossed bool) {
			lsb := byte(operand) + s.X
			msb := lsb + 1
			return *s.Memory[(uint16(*s.Memory[msb])<<8)|uint16(*s.Memory[lsb])], false
		}
	case AddressIndirectY:
		return func(s *State) (value byte, pageCrossed bool) {
			lsb := *s.Memory[byte(operand)]
			msb := *s.Memory[byte(operand)+1]
			return *s.Memory[((uint16(msb)<<8)|uint16(*s.Memory[lsb]))+uint16(s.Y)], false
		}
	case AddressRelative:
		addr, c := calculateRelativeAddress(instructionLength, operand, address)
		return func(s *State) (value byte, pageCrossed bool) { return *s.Memory[addr], c }
	default:
		return func(s *State) (value byte, pageCrossed bool) { return 0, false }

	}
}

func getSetter(addressMode int, address, instructionLength, operand uint16) Setter {
	switch addressMode {
	case AddressAccumulator:
		return func(s *State, value byte) bool {
			s.A = value
			return false
		}
	case AddressZeroPage:
		addr := byte(operand)
		return func(s *State, value byte) bool {
			*s.Memory[addr] = value
			return false
		}
	case AddressAbsolute, AddressAddress:
		return func(s *State, value byte) bool {
			*s.Memory[operand] = value
			return false
		}
	case AddressZeroPageX:
		return func(s *State, value byte) bool {
			*s.Memory[byte(operand)+s.X] = value
			return false
		}
	case AddressZeroPageY:
		return func(s *State, value byte) bool {
			*s.Memory[byte(operand)+s.Y] = value
			return false
		}
	case AddressAbsoluteX:
		return func(s *State, value byte) bool {
			*s.Memory[operand+uint16(s.X)] = value
			return false
		}
	case AddressAbsoluteY:
		return func(s *State, value byte) bool {
			*s.Memory[operand+uint16(s.Y)] = value
			return false
		}
	case AddressIndirect:
		return func(s *State, value byte) bool {
			*s.Memory[(uint16(*s.Memory[operand+1])<<8)|uint16(*s.Memory[operand])] = value
			return false
		}
	case AddressIndirectX:
		return func(s *State, value byte) bool {
			lsb := byte(operand) + s.X
			msb := lsb + 1
			*s.Memory[(uint16(*s.Memory[msb])<<8)|uint16(*s.Memory[lsb])] = value
			return false
		}
	case AddressIndirectY:
		return func(s *State, value byte) bool {
			lsb := *s.Memory[byte(operand)]
			msb := *s.Memory[byte(operand)+1]
			*s.Memory[((uint16(msb)<<8)|uint16(*s.Memory[lsb]))+uint16(s.Y)] = value
			return false
		}
	case AddressRelative:
		addr, c := calculateRelativeAddress(instructionLength, operand, address)
		return func(s *State, value byte) bool {
			*s.Memory[addr] = value
			return c
		}
	default:
		return func(s *State, value byte) bool { return false }
	}
}

func (s *State) GetValue(addressMode int, instructionLength, offset uint16) (value byte, pageCrossed bool) {
	switch addressMode {
	case AddressImmediate:
		return byte(offset + instructionLength), false
	case AddressAccumulator:
		return s.A, false
	case AddressImplied:
		return 0, false // Not sure why we should ever get here
	default:
		a, c := s.CalculateAddress(addressMode, instructionLength, offset)
		if a >= 0 && a < uint16(len(s.Memory)) {
			return *s.Memory[a], c
		} else {
			return 0, c
		}
	}
}
