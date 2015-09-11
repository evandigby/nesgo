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

type Registers struct {
	PC uint16
	A  uint8
	X  uint8
	Y  uint8
	SP uint8

	Status *Flags
}

type State struct {
	*Registers

	Memory []*byte
}

func NewState() *State {
	return &State{&Registers{Status: &Flags{}}, []*byte{}}
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
	s.Status.Interrupt = true
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

func (s *State) calculateAddress(addressMode int, offset int16) int {
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

	return 0
}

func (s *State) GetValue(addressMode int, offset int16) byte {
	switch addressMode {
	case AddressImmediate:
		return byte(offset)
	case AddressAccumulator:
		return byte(s.A)
	case AddressImplied:
		return byte(0) // Not sure why we should ever get here
	default:
		return *s.Memory[s.calculateAddress(addressMode, address, offset)]
	}

	return byte(0)
}

func (s *State) SetValue(addressMode int, address, offset int16, val byte) {
	*s.Memory[s.calculateAddress(addressMode, address, offset)] = val
}
