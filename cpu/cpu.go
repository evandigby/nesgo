package cpu

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/evandigby/nesgo/nes"
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
	nes *nes.NES `json:"-"`
	Registers
	Flags

	Executers []Executer `json:"-"`
	Opcodes   []*Opcode  `json:"-"`

	nintendulatorLog bool          `json:"-"`
	cpuLog           *os.File      `json:"-"`
	nesLog           *bufio.Reader `json:"-"`
	Sync             chan int      `json:"-"`
	exit             chan bool     `json:"-"`
}

func NewCPU(n *nes.NES, exit chan bool, log, nesLog *os.File) *CPU {
	return &CPU{
		nes:    n,
		cpuLog: log,
		nesLog: bufio.NewReader(nesLog),
		Sync:   make(chan int),
		exit:   exit,
	}
}

func (c *CPU) loadOpcodes() {
	c.Opcodes = make([]*Opcode, len(c.nes.Memory))
	c.Executers = make([]Executer, len(c.nes.Memory))
	for i, _ := range c.nes.Memory {
		op := NewOpcode(c.nes.Memory, uint16(i))
		if op != nil {
			c.Opcodes[i] = op
			c.Executers[i] = op.Executer()
		}
	}
}

func (c *CPU) execute() {
	c.PowerUp()
	c.loadOpcodes()
	//c.PC = 0xC000

	cs := 0
	instructionsRun := 0
	c.nintendulatorLog = false
	for {
		<-c.Sync
		if c.nintendulatorLog {
			c.nes.Debug = true
			op := c.Opcodes[c.PC]
			disassembly := fmt.Sprintf("%v %v", op.Disassemble(), op.GetValueAt(c))
			ppuc := (cs * 3) % 341
			log := fmt.Sprintf("%04X  %-8s %-32s A:%02X X:%02X Y:%02X P:%02X SP:%02X CYC:%3s\n", c.PC, op.Bytes(), disassembly, c.A, c.X, c.Y, c.Status(), c.SP, strconv.FormatInt(int64(ppuc), 10))
			fmt.Printf("%v: %v", instructionsRun, log)

			if c.cpuLog != nil {
				c.cpuLog.WriteString(log)
			}

			/*
				if c.nesLog != nil {
					l, _, err := c.nesLog.ReadLine()
					if err != nil {
						panic(err)
					}
					lastLogLine := nesLogLine
					nesLogLine = string(l[0:81])
						if nesLogLine != log[0:81] {
							fmt.Printf("%v: %v\n", instructionsRun-1, lastLogLine)
							fmt.Printf("%v: %v\n", instructionsRun, nesLogLine)

							c.exit <- true
						}
				}
			*/
			instructionsRun++
			c.nes.Debug = false
			//			if instructionsRun >= 8991 {
		}
		cycles := c.Execute()
		cs += cycles

		c.Sync <- cycles

	}
}

func (c *CPU) Run() {
	go c.execute()
}

func (c *CPU) Execute() int {
	cycles, pc := c.Executers[c.PC](c)
	c.PC = pc

	return cycles
}

func (c *CPU) Push(val byte) {
	*c.nes.Stack[c.SP] = val
	c.SP--
}

func (c *CPU) Pop() byte {
	c.SP++
	return *c.nes.Stack[c.SP]
}

func (c *CPU) PowerUp() {
	c.PC = uint16(*c.nes.Memory[0xFFFC]) | (uint16(*c.nes.Memory[0xFFFD]) << 8)
	c.A = 0
	c.X = 0
	c.Y = 0
	c.SP = 0xFD
	c.Interrupt = true
}

func (c *CPU) Reset() {
	c.PC = uint16(*c.nes.Memory[0xFFFC]) | (uint16(*c.nes.Memory[0xFFFD]) << 8)
	c.SP -= 3
	c.Interrupt = true
	*c.nes.Memory[0x4015] = 0x00
}

func calculateRelativeAddress(instructionLength, offset, pc uint16) (uint16, bool) {
	cmp := (pc + instructionLength) & 0xFF00
	if offset&0x80 == 0 {
		addr := pc + instructionLength + (offset & 0x7F)
		return addr, cmp != (addr & 0xFF00)
	} else {
		addr := pc + instructionLength - (0x80 - (offset & 0x7F))
		return addr, cmp != (addr & 0xFF00)
	}
}

func intermediateAddressCalculator(addressMode int, instructionLength, offset uint16) AddressGetter {
	switch addressMode {
	case AddressIndirect:
		return func(c *CPU) (uint16, bool) { return offset, false }
	case AddressIndirectX:
		return func(c *CPU) (uint16, bool) { return uint16(byte(offset) + c.X), false }
	case AddressIndirectY:
		return func(c *CPU) (uint16, bool) {
			lsb := uint16(*c.nes.Memory[byte(offset)])
			msb := uint16(*c.nes.Memory[byte(offset)+1])
			return ((msb << 8) | lsb), false
		}
	}

	return func(c *CPU) (uint16, bool) { return 0, false }
}

type Getter func(c *CPU) (value byte, pageCrossed bool)
type AddressGetter func(c *CPU) (uint16, bool)
type Setter func(c *CPU, value byte) bool

func addressCalculator(addressMode int, address, instructionLength, operand uint16) AddressGetter {
	switch addressMode {
	case AddressZeroPage:
		return func(c *CPU) (uint16, bool) { return uint16(byte(operand)), false }
	case AddressAbsolute, AddressAddress:
		return func(c *CPU) (uint16, bool) {
			return operand, false
		}
	case AddressZeroPageX:
		return func(c *CPU) (uint16, bool) { return uint16(byte(operand) + c.X), false }
	case AddressZeroPageY:
		return func(c *CPU) (uint16, bool) { return uint16(byte(operand) + c.Y), false }
	case AddressAbsoluteX:
		return func(c *CPU) (uint16, bool) {
			return operand + uint16(c.X), (operand&0x00FF)+uint16(c.X) > 0xFF
		}
	case AddressAbsoluteY:
		return func(c *CPU) (uint16, bool) { return operand + uint16(c.Y), (operand&0x00FF)+uint16(c.Y) > 0xFF }
	case AddressIndirect:
		lsb := operand
		msb := uint16(byte(operand)+byte(1)) | (operand & 0xFF00)
		return func(c *CPU) (uint16, bool) {
			return (uint16(*c.nes.Memory[msb]) << 8) | uint16(*c.nes.Memory[lsb]), false
		}
	case AddressIndirectWrong:
		lsb := operand
		msb := operand + 1
		return func(c *CPU) (uint16, bool) {
			return (uint16(*c.nes.Memory[msb]) << 8) | uint16(*c.nes.Memory[lsb]), false
		}
	case AddressIndirectX:
		return func(c *CPU) (uint16, bool) {
			lsb := byte(operand) + c.X
			msb := lsb + 1
			return (uint16(*c.nes.Memory[msb]) << 8) | uint16(*c.nes.Memory[lsb]), false
		}
	case AddressIndirectY:
		return func(c *CPU) (uint16, bool) {
			lsb := uint16(*c.nes.Memory[byte(operand)])
			msb := uint16(*c.nes.Memory[byte(operand)+1])
			return ((msb << 8) | lsb) + uint16(c.Y), lsb+uint16(c.Y) > 0x100
		}
	case AddressRelative:
		addr, _ := calculateRelativeAddress(instructionLength, operand, address)
		return func(c *CPU) (uint16, bool) { return addr, false }
	default:
		return func(c *CPU) (uint16, bool) { return 0, false }
	}
}

func getGetter(addressMode int, address, instructionLength, operand uint16) Getter {
	ac := addressCalculator(addressMode, address, instructionLength, operand)

	switch addressMode {
	case AddressImmediate:
		return func(c *CPU) (value byte, pageCrossed bool) { return byte(operand), false }
	case AddressAccumulator:
		return func(c *CPU) (value byte, pageCrossed bool) { return c.A, false }
	case AddressImplied:
		return func(c *CPU) (value byte, pageCrossed bool) { return 0, false } // Not sure why we should ever get here
	case AddressZeroPage:
		addr, _ := ac(nil)
		return func(c *CPU) (value byte, pageCrossed bool) {
			return c.nes.Get(addr), false
		}
	case AddressAbsolute, AddressAddress:
		addr, _ := ac(nil)
		return func(c *CPU) (value byte, pageCrossed bool) {
			return c.nes.Get(addr), false
		}
	default:
		return func(c *CPU) (value byte, pageCrossed bool) {
			addr, cycles := ac(c)
			return c.nes.Get(addr), cycles
		}
	}
}

func getSetter(addressMode int, address, instructionLength, operand uint16) Setter {
	ac := addressCalculator(addressMode, address, instructionLength, operand)

	switch addressMode {
	case AddressAccumulator:
		return func(c *CPU, value byte) bool {
			c.A = value
			return false
		}
	case AddressZeroPage:
		addr, _ := ac(nil)

		return func(c *CPU, value byte) bool {
			c.nes.Set(addr, value)
			c.invalidateExecutor(addr)
			return false
		}
	default:
		return func(c *CPU, value byte) bool {
			addr, cycles := ac(c)
			c.nes.Set(addr, value)
			c.invalidateExecutor(addr)
			return cycles
		}
	}
}

func (c *CPU) invalidateExecutor(address uint16) {
	start := address - 2
	if start < 0 {
		start = 0
	}

	for i := start; i <= address; i++ {
		op := NewOpcode(c.nes.Memory, i)
		c.Opcodes[i] = op
		c.Executers[i] = op.Executer()
	}
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
