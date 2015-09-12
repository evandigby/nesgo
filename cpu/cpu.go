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

	Memory []*byte
	Stack  []*byte
}

func (s *State) Push(val byte) {
	*s.Stack[s.SP] = val
	s.SP++
}

func (s *State) Pop() byte {
	s.SP--
	return *s.Stack[s.SP]
}

func NewState() *State {
	m := make([]*byte, 0x8000)
	s := m[0x0100:0x01FF]
	return &State{Registers{}, Flags{}, m, s}
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

func (s *State) CalculateAddress(addressMode int, offset uint16) (address uint16, addedCycles int) {
	return 0, 0 /*
		switch addressMode {
		case AddressZeroPage, AddressAbsolute:
			return offset
		case AddressZeroPageX, AddressAbsoluteX:
			return offset + s.X
		case AddressZeroPageY, AddressAbsoluteY:
			return offset + s.Y
		case AddressIndirect:
			return int((uint16(s.Memory[offset+1]) << 8) | uint16(s.Memory[offset]))
		case AddressIndirectX:
			offset += s.X
			return int((uint16(s.Memory[offset+1]) << 8) | uint16(s.Memory[offset]))
		case AddressIndirectY:
			return int((uint16(s.Memory[offset+1])<<8)|uint16(s.Memory[offset])) + int(s.Y)
		case AddressRelative:
			if offset&0x80 != 0 {
				return int(s.PC + (offset & 0x7F))
			} else {
				return int(s.PC - (offset & 0x7F))
			}
		}

		return 0*/
}

func (s *State) GetValue(addressMode int, offset uint16) (value byte, addedCycles int) {
	switch addressMode {
	case AddressImmediate:
		return byte(offset), 0
	case AddressAccumulator:
		return s.A, 0
	case AddressImplied:
		return 0, 0 // Not sure why we should ever get here
	default:
		a, c := s.CalculateAddress(addressMode, offset)
		return *s.Memory[a], c
	}
}

func (s *State) SetValue(addressMode int, offset uint16, val byte) int {
	switch addressMode {
	case AddressAccumulator:
		s.A = val
		return 0
	default:
		v, c := s.CalculateAddress(addressMode, offset)
		*s.Memory[v] = val
		return c
	}
}
