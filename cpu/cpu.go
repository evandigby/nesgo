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

func calculateAddress(addressMode int, address, offset int16) int {
	switch addressMode {
	case AddressZeroPage:
	case AddressZeroPageX:
	case AddressZeroPageY:
	case AddressAbsolute:
	case AddressAbsoluteX:
	case AddressAbsoluteY:
	case AddressAccumulator:
	case AddressIndirect:
	case AddressIndirectX:
	case AddressIndirectY:
	case AddressRelative:

	}

	return 0
}

func (s *State) GetValue(addressMode int, address, offset int16) byte {
	switch addressMode {
	case AddressImmediate:
		return byte(offset)
	case AddressImplied:
		return byte(0) // Not sure why we should ever get here
	default:
		return *s.Memory[calculateAddress(addressMode, address, offset)]
	}

	return byte(0)
}

func (s *State) SetValue(addressMode int, address, offset int16, val byte) {
	*s.Memory[calculateAddress(addressMode, address, offset)] = val
}
