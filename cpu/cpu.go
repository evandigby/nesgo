package cpu

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

	Memory         []*byte
	Stack          []*byte
	PPURegisters   []*byte
	APUIORegisters []*byte
	Cartridge      []*byte
}

func NewState() *State {
	tm := make([]byte, 0xFFFF)
	m := make([]*byte, 0xFFFF)
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
	s := m[0x0100:0x01FF]

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
	c := m[0x4020:0xFFFF]
	return &State{Registers{}, Flags{}, m, s, ppu, apu, c}
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

func (s *State) CalculateAddress(addressMode int, offset uint16) (address uint16, pageCrossed bool) {
	switch addressMode {
	case AddressZeroPage, AddressAbsolute:
		return offset, false
	case AddressZeroPageX, AddressAbsoluteX:
		return offset + uint16(s.X), false
	case AddressZeroPageY, AddressAbsoluteY:
		return offset + uint16(s.Y), false
	case AddressIndirect:
		return (uint16(*s.Memory[offset+1]) << 8) | uint16(*s.Memory[offset]), false
	case AddressIndirectX:
		offset += uint16(s.X)
		return (uint16(*s.Memory[offset+1]) << 8) | uint16(*s.Memory[offset]), false
	case AddressIndirectY:
		return (uint16(*s.Memory[offset+1]) << 8) | uint16(*s.Memory[offset]) + uint16(s.Y), false
	case AddressRelative:
		if offset&0x80 != 0 {
			return s.PC + (offset & 0x7F), false
		} else {
			return s.PC - (offset & 0x7F), false
		}
	}

	return 0, false
}

func (s *State) GetValue(addressMode int, offset uint16) (value byte, pageCrossed bool) {
	switch addressMode {
	case AddressImmediate:
		return byte(offset), false
	case AddressAccumulator:
		return s.A, false
	case AddressImplied:
		return 0, false // Not sure why we should ever get here
	default:
		a, c := s.CalculateAddress(addressMode, offset)
		return *s.Memory[a], c
	}
}

func (s *State) SetValue(addressMode int, offset uint16, val byte) bool {
	switch addressMode {
	case AddressAccumulator:
		s.A = val
		return false
	default:
		v, c := s.CalculateAddress(addressMode, offset)
		*s.Memory[v] = val
		return c
	}
}
