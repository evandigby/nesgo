package cpu

import (
	"bufio"
	"fmt"
	"os"
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
	AddressIndirectWrong
	AddressIndirectX
	AddressIndirectY
	AddressRelative
	AddressAddress
)

type CPU struct {
	State *State
	rom   rom.ROM

	nintendulatorLog bool
	cpuLog           *os.File
	nesLog           *bufio.Reader
	Sync             chan int
	exit             chan bool
}

func NewCPU(r rom.ROM, exit chan bool, log, nesLog *os.File) *CPU {
	return &CPU{NewState(), r, true, log, bufio.NewReader(nesLog), make(chan int), exit}
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

func (c *CPU) execute() {
	c.loadRom(c.rom)
	c.State.PowerUp()
	c.State.PC = 0xC000

	cs := 0
	instructionsRun := 0
	var nesLogLine string
	for {
		<-c.Sync
		if c.nintendulatorLog {
			op := c.State.Opcodes[c.State.PC]
			disassembly := fmt.Sprintf("%v %v", op.Disassemble(), op.GetValueAt(c.State))
			ppuc := (cs * 3) % 341
			log := fmt.Sprintf("%04X  %-8s %-32s A:%02X X:%02X Y:%02X P:%02X SP:%02X CYC:%3s\n", c.State.PC, op.Bytes(), disassembly, c.State.A, c.State.X, c.State.Y, c.State.Status(), c.State.SP, strconv.FormatInt(int64(ppuc), 10))
			fmt.Printf("%v: %v", instructionsRun, log)
			if c.cpuLog != nil {
				c.cpuLog.WriteString(log)
			}

			if c.nesLog != nil {
				l, _, err := c.nesLog.ReadLine()
				if err != nil {
					panic(err)
				}
				lastLogLine := nesLogLine
				nesLogLine = string(l[0:73])
				if nesLogLine != log[0:73] {
					fmt.Printf("%v: %v\n", instructionsRun-1, lastLogLine)
					fmt.Printf("%v: %v\n", instructionsRun, nesLogLine)
					c.exit <- true
				}
			}
			instructionsRun++
			if instructionsRun >= 8991 {
				c.exit <- true
			}
		}
		cycles := c.State.Execute()
		cs += cycles

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
	Negative  bool
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
	if f.Negative {
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
	f.Negative = v&128 != 0
}

func (f *Flags) SetSign(value byte) {
	f.Negative = value&0x80 != 0
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

const MemSize = 0xFFFF

func NewState() *State {
	tm := make([]byte, MemSize+1)
	m := make([]*byte, MemSize+1)
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
	c := m[0x8000 : MemSize+1]

	return &State{Registers{}, Flags{}, m, s, ppu, apu, c, nil, nil}
}

func (s *State) Execute() int {
	cycles, pc := s.Executers[s.PC](s)
	s.PC = pc

	return cycles
}

func (s *State) Push(val byte) {
	*s.Stack[s.SP] = val
	s.SP--
}

func (s *State) Pop() byte {
	s.SP++
	return *s.Stack[s.SP]
}

func (s *State) PowerUp() {
	s.PC = uint16(*s.Memory[0xFFFC]) | (uint16(*s.Memory[0xFFFD]) << 8)
	s.A = 0
	s.X = 0
	s.Y = 0
	s.SP = 0xFD
	s.Interrupt = true
}

func (s *State) Reset() {
	s.PC = uint16(*s.Memory[0xFFFC]) | (uint16(*s.Memory[0xFFFD]) << 8)
	s.SP -= 3
	s.Interrupt = true
}

func calculateRelativeAddress(instructionLength, offset, pc uint16) (uint16, bool) {
	if offset&0x80 == 0 {
		return pc + instructionLength + (offset & 0x7F), false
	} else {
		return pc + instructionLength - (0x80 - (offset & 0x7F)), false
	}
}

func intermediateAddressCalculator(addressMode int, instructionLength, offset uint16) AddressGetter {
	switch addressMode {
	case AddressIndirect:
		return func(s *State) uint16 { return offset }
	case AddressIndirectX:
		return func(s *State) uint16 { return uint16(byte(offset) + s.X) }
	case AddressIndirectY:
		return func(s *State) uint16 {
			lsb := uint16(*s.Memory[byte(offset)])
			msb := uint16(*s.Memory[byte(offset)+1])
			return ((msb << 8) | lsb)
		}
	}

	return func(s *State) uint16 { return 0 }
}

func addressCalculator(addressMode int, address, instructionLength, operand uint16) AddressGetter {
	switch addressMode {
	case AddressZeroPage:
		return func(s *State) uint16 { return uint16(byte(operand)) }
	case AddressAbsolute, AddressAddress:
		return func(s *State) uint16 {
			return operand
		}
	case AddressZeroPageX:
		return func(s *State) uint16 { return uint16(byte(operand) + s.X) }
	case AddressZeroPageY:
		return func(s *State) uint16 { return uint16(byte(operand) + s.Y) }
	case AddressAbsoluteX:
		return func(s *State) uint16 { return operand + uint16(s.X) }
	case AddressAbsoluteY:
		return func(s *State) uint16 { return operand + uint16(s.Y) }
	case AddressIndirect:
		lsb := operand
		msb := uint16(byte(operand)+byte(1)) | (operand & 0xFF00)
		return func(s *State) uint16 {
			return (uint16(*s.Memory[msb]) << 8) | uint16(*s.Memory[lsb])
		}
	case AddressIndirectWrong:
		lsb := operand
		msb := operand + 1
		return func(s *State) uint16 {
			return (uint16(*s.Memory[msb]) << 8) | uint16(*s.Memory[lsb])
		}
	case AddressIndirectX:
		return func(s *State) uint16 {
			lsb := byte(operand) + s.X
			msb := lsb + 1
			return (uint16(*s.Memory[msb]) << 8) | uint16(*s.Memory[lsb])
		}
	case AddressIndirectY:
		return func(s *State) uint16 {
			lsb := uint16(*s.Memory[byte(operand)])
			msb := uint16(*s.Memory[byte(operand)+1])
			return ((msb << 8) | lsb) + uint16(s.Y)
		}
	case AddressRelative:
		addr, _ := calculateRelativeAddress(instructionLength, operand, address)
		return func(s *State) uint16 { return addr }
	default:
		return func(s *State) uint16 { return 0 }
	}
}

type Getter func(s *State) (value byte, pageCrossed bool)
type AddressGetter func(s *State) uint16
type Setter func(s *State, value byte) bool

func getGetter(addressMode int, address, instructionLength, operand uint16) Getter {
	ac := addressCalculator(addressMode, address, instructionLength, operand)

	switch addressMode {
	case AddressImmediate:
		return func(s *State) (value byte, pageCrossed bool) { return byte(operand), false }
	case AddressAccumulator:
		return func(s *State) (value byte, pageCrossed bool) { return s.A, false }
	case AddressImplied:
		return func(s *State) (value byte, pageCrossed bool) { return 0, false } // Not sure why we should ever get here
	case AddressZeroPage:
		addr := ac(nil)
		return func(s *State) (value byte, pageCrossed bool) { return *s.Memory[addr], false }
	case AddressAbsolute, AddressAddress:
		addr := ac(nil)
		return func(s *State) (value byte, pageCrossed bool) { return *s.Memory[addr], false }
	default:
		return func(s *State) (value byte, pageCrossed bool) {
			addr := ac(s)
			return *s.Memory[addr], false
		}
	}
}

func getSetter(addressMode int, address, instructionLength, operand uint16) Setter {
	ac := addressCalculator(addressMode, address, instructionLength, operand)

	switch addressMode {
	case AddressAccumulator:
		return func(s *State, value byte) bool {
			s.A = value
			return false
		}
	case AddressZeroPage:
		addr := ac(nil)

		return func(s *State, value byte) bool {
			*s.Memory[addr] = value
			s.invalidateExecutor(addr)
			return false
		}
	default:
		return func(s *State, value byte) bool {
			addr := ac(s)
			*s.Memory[addr] = value
			s.invalidateExecutor(addr)
			return false
		}
		/*
			case AddressAbsolute, AddressAddress:
				return func(s *State, value byte) bool {
					*s.Memory[operand] = value
					s.invalidateExecutor(operand)
					return false
				}
			case AddressZeroPageX:
				return func(s *State, value byte) bool {
					addr := uint16(byte(operand) + s.X)
					*s.Memory[addr] = value
					s.invalidateExecutor(addr)
					return false
				}
			case AddressZeroPageY:
				return func(s *State, value byte) bool {
					addr := uint16(byte(operand) + s.Y)
					*s.Memory[addr] = value
					s.invalidateExecutor(addr)
					return false
				}
			case AddressAbsoluteX:
				return func(s *State, value byte) bool {
					addr := operand + uint16(s.X)
					*s.Memory[addr] = value
					s.invalidateExecutor(addr)
					return false
				}
			case AddressAbsoluteY:
				return func(s *State, value byte) bool {
					addr := operand + uint16(s.Y)
					*s.Memory[addr] = value
					s.invalidateExecutor(addr)
					return false
				}
			case AddressIndirect:
				lsb := operand
				msb := uint16((byte(operand) + byte(1))) | (operand & 0xFF00)

				return func(s *State, value byte) bool {
					addr := (uint16(*s.Memory[msb]) << 8) | uint16(*s.Memory[lsb])
					*s.Memory[addr] = value
					s.invalidateExecutor(addr)
					return false
				}
			case AddressIndirectX:
				return func(s *State, value byte) bool {
					lsb := byte(operand) + s.X
					msb := lsb + 1
					addr := (uint16(*s.Memory[msb]) << 8) | uint16(*s.Memory[lsb])
					*s.Memory[addr] = value
					s.invalidateExecutor(addr)
					return false
				}
			case AddressIndirectY:
				return func(s *State, value byte) bool {
					lsb := *s.Memory[byte(operand)]
					msb := *s.Memory[byte(operand)+1]
					addr := ((uint16(msb) << 8) | uint16(*s.Memory[lsb])) + uint16(s.Y)
					*s.Memory[addr] = value
					s.invalidateExecutor(addr)
					return false
				}
			case AddressRelative:
				addr, c := calculateRelativeAddress(instructionLength, operand, address)
				return func(s *State, value byte) bool {
					*s.Memory[addr] = value
					s.invalidateExecutor(addr)
					return c
				}
			default:
				return func(s *State, value byte) bool { return false }
		*/
	}
}

func (s *State) invalidateExecutor(address uint16) {
	start := address - 2
	end := address + 2
	if start < 0 {
		start = 0
	}
	if end > MemSize {
		end = MemSize
	}

	for i := start; i < end; i++ {
		op := NewOpcode(s.Memory, i)
		s.Opcodes[i] = op
		s.Executers[i] = op.Executer()
	}
}
