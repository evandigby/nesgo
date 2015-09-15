package cpu

import "github.com/evandigby/nesgo/rom"

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
)

type CPU struct {
	State *State
	rom   rom.ROM

	Sync chan int
}

func NewCPU(r rom.ROM) *CPU {
	return &CPU{NewState(), r, make(chan int)}
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
		op := NewOpcode(c.State.Memory, i)
		if op != nil {
			c.State.Opcodes[i] = op
			c.State.Executers[i] = op.Executer()
		}
	}
}

func (c *CPU) execute() {
	c.loadRom(c.rom)
	c.State.PC = 0xC000

	for {
		<-c.Sync
		cycles, pc := c.State.Executers[c.State.PC](c.State)
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
	s := m[0x0100:0xFFFF] //m[0x0100:0x01FF]

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
	apu := m[0x4000:0x401F]

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
}

func (s *State) Reset() {
	s.SP -= 3
	s.Interrupt = true
}

func (s *State) CalculateAddress(addressMode int, instructionLength, offset uint16) (address uint16, pageCrossed bool) {
	switch addressMode {
	case AddressZeroPage:
		return uint16(byte(offset)), false
	case AddressAbsolute:
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
		if offset&0x80 == 0 {
			return s.PC + instructionLength + (offset & 0x7F), false
		} else {
			return s.PC - (offset & 0x7F), false
		}
	}

	return 0, false
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

func (s *State) SetValue(addressMode int, instructionLength, offset uint16, val byte) bool {
	switch addressMode {
	case AddressAccumulator:
		s.A = val
		return false
	default:
		v, c := s.CalculateAddress(addressMode, instructionLength, offset)
		if v >= 0 && v < uint16(len(s.Memory)) {
			*s.Memory[v] = val
		}
		return c
	}
}
