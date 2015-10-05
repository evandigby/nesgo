// Decimal Mode Not Implemented -- doesn't exist in NES
package cpu

func adc(c *CPU, v byte) {
	t := uint16(v) + uint16(c.A)
	if c.Carry {
		t += 1
	}
	bt := byte(t)
	c.Carry = t > 0xFF
	c.Overflow = ((((c.A ^ v) & 0x80) == 0) && ((c.A^bt)&0x80) != 0)
	c.SetZero(bt)
	c.SetSign(bt)
	c.A = bt
}

func ADC(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, pageCrossed := get(c)
		adc(c, v)
		if pageCrossed {
			return cycles + 1, c.PC + instructionLength
		} else {
			return cycles, c.PC + instructionLength
		}
	}
}

func and(c *CPU, v byte) {
	v = v & c.A
	c.SetSign(v)
	c.SetZero(v)
	c.A = byte(v)
}

func AND(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, pageCrossed := get(c)
		and(c, v)

		if pageCrossed {
			return cycles + 1, c.PC + instructionLength
		} else {
			return cycles, c.PC + instructionLength
		}
	}
}

func SAX(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v := c.A & c.X
		//c.SetSign(c.A)
		//c.SetZero(c.A)
		set(c, v)
		return cycles, c.PC + instructionLength
	}
}

func asl(c *CPU, v byte) byte {
	c.Carry = v&0x80 != 0
	v <<= 1
	v &= 0xFF
	c.SetSign(v)
	c.SetZero(v)
	return v
}

func ASL(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, pageCrossed := get(c)

		v = asl(c, v)

		set(c, v)
		if pageCrossed {
			return cycles + 1, c.PC + instructionLength
		} else {
			return cycles, c.PC + instructionLength
		}
	}
}

func branch(b bool, c *CPU, v uint16, pageCross bool, instructionLength uint16, operand uint16, cycles int) (int, uint16) {
	if b {
		if pageCross {
			cycles += 2
		} else {
			cycles++
		}

		return cycles, v
	} else {
		return cycles, c.PC + instructionLength
	}
}

func BCC(address, instructionLength, operand uint16, cycles int) Executer {
	v, pageCrossed := calculateRelativeAddress(instructionLength, operand, address)
	return func(c *CPU) (int, uint16) {
		return branch(!c.Carry, c, v, pageCrossed, instructionLength, operand, cycles)
	}
}
func BCS(address, instructionLength, operand uint16, cycles int) Executer {
	v, pageCrossed := calculateRelativeAddress(instructionLength, operand, address)
	return func(c *CPU) (int, uint16) {
		return branch(c.Carry, c, v, pageCrossed, instructionLength, operand, cycles)
	}
}
func BEQ(address, instructionLength, operand uint16, cycles int) Executer {
	v, pageCrossed := calculateRelativeAddress(instructionLength, operand, address)
	return func(c *CPU) (int, uint16) {
		return branch(c.Zero, c, v, pageCrossed, instructionLength, operand, cycles)
	}
}
func BIT(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, pageCrossed := get(c)

		r := c.A & byte(v)
		c.SetZero(r)
		c.Overflow = byte(v)&0x40 != 0
		c.SetSign(byte(v))

		if pageCrossed {
			return cycles + 1, c.PC + instructionLength
		} else {
			return cycles, c.PC + instructionLength
		}
	}
}
func BMI(address, instructionLength, operand uint16, cycles int) Executer {
	v, pageCrossed := calculateRelativeAddress(instructionLength, operand, address)
	return func(c *CPU) (int, uint16) {
		return branch(c.Negative, c, v, pageCrossed, instructionLength, operand, cycles)
	}
}

func BNE(address, instructionLength, operand uint16, cycles int) Executer {
	v, pageCrossed := calculateRelativeAddress(instructionLength, operand, address)
	return func(c *CPU) (int, uint16) {
		return branch(!c.Zero, c, v, pageCrossed, instructionLength, operand, cycles)
	}
}
func BPL(address, instructionLength, operand uint16, cycles int) Executer {
	v, pageCrossed := calculateRelativeAddress(instructionLength, operand, address)
	return func(c *CPU) (int, uint16) {
		return branch(!c.Negative, c, v, pageCrossed, instructionLength, operand, cycles)
	}
}
func BRK(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		c.PC++
		c.Push(byte(c.PC >> 8))
		c.Push(byte(c.PC))
		c.Break = true
		c.Push(c.Status())
		c.Interrupt = true
		lsd := c.nes.Get(0xFFFE)
		msd := c.nes.Get(0xFFFF)
		return cycles, uint16(lsd) | (uint16(msd) << 8)
	}
}
func BVC(address, instructionLength, operand uint16, cycles int) Executer {
	v, pageCrossed := calculateRelativeAddress(instructionLength, operand, address)
	return func(c *CPU) (int, uint16) {
		return branch(!c.Overflow, c, v, pageCrossed, instructionLength, operand, cycles)
	}
}
func BVS(address, instructionLength, operand uint16, cycles int) Executer {
	v, pageCrossed := calculateRelativeAddress(instructionLength, operand, address)
	return func(c *CPU) (int, uint16) {
		return branch(c.Overflow, c, v, pageCrossed, instructionLength, operand, cycles)
	}
}
func CLC(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		c.Carry = false
		return cycles, c.PC + instructionLength
	}
}
func CLD(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		c.Decimal = false
		return cycles, c.PC + instructionLength
	}
}
func CLI(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		c.Interrupt = false
		return cycles, c.PC + instructionLength
	}
}
func CLV(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		c.Overflow = false
		return cycles, c.PC + instructionLength
	}
}

func compare(v byte, pageCrossed bool, instructionLength uint16, operand uint16, cycles int, V byte, c *CPU) (int, uint16) {
	t := V - v

	c.Carry = V >= v
	c.Zero = V == v
	c.SetSign(t)
	if pageCrossed {
		return cycles + 1, c.PC + instructionLength
	} else {
		return cycles, c.PC + instructionLength
	}
}

func CMP(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, pageCrossed := get(c)
		return compare(v, pageCrossed, instructionLength, operand, cycles, c.A, c)
	}
}
func CPX(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, pageCrossed := get(c)
		return compare(v, pageCrossed, instructionLength, operand, cycles, c.X, c)
	}
}
func CPY(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, pageCrossed := get(c)
		return compare(v, pageCrossed, instructionLength, operand, cycles, c.Y, c)
	}
}
func DEC(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, _ := get(c)
		v -= 1
		c.SetZero(v)
		c.SetSign(v)
		set(c, v)
		return cycles, c.PC + instructionLength
	}
}
func DEX(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		c.X -= 1
		c.SetZero(c.X)
		c.SetSign(c.X)
		return cycles, c.PC + instructionLength
	}
}
func DEY(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		c.Y -= 1
		c.SetZero(c.Y)
		c.SetSign(c.Y)
		return cycles, c.PC + instructionLength
	}
}

func eor(c *CPU, v byte) {
	c.A = c.A ^ v
	c.SetZero(c.A)
	c.SetSign(c.A)
}

func EOR(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, pageCrossed := get(c)
		eor(c, v)
		if pageCrossed {
			return cycles + 1, c.PC + instructionLength
		} else {
			return cycles, c.PC + instructionLength
		}
	}
}
func INC(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, _ := get(c)
		v += 1
		c.SetZero(v)
		c.SetSign(v)
		set(c, v)
		return cycles, c.PC + instructionLength
	}
}
func INX(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		c.X += 1
		c.SetZero(c.X)
		c.SetSign(c.X)
		return cycles, c.PC + instructionLength
	}
}
func INY(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		c.Y += 1
		c.SetZero(c.Y)
		c.SetSign(c.Y)
		return cycles, c.PC + instructionLength
	}
}
func JMP(get AddressGetter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		a, _ := get(c)
		return cycles, a
	}
}
func JSR(get AddressGetter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v := c.PC + instructionLength - 1
		c.Push(byte(v >> 8))
		c.Push(byte(v))
		a, _ := get(c)
		return cycles, a
	}
}
func LDA(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, pageCrossed := get(c)
		c.A = v
		c.SetZero(v)
		c.SetSign(v)
		if pageCrossed {
			return cycles + 1, c.PC + instructionLength
		} else {
			return cycles, c.PC + instructionLength
		}
	}
}
func LAX(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, pageCrossed := get(c)
		c.A = v
		c.X = v
		c.SetZero(v)
		c.SetSign(v)
		if pageCrossed {
			return cycles + 1, c.PC + instructionLength
		} else {
			return cycles, c.PC + instructionLength
		}
	}
}
func LDX(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, pageCrossed := get(c)
		c.X = v
		c.SetZero(v)
		c.SetSign(v)
		if pageCrossed {
			return cycles + 1, c.PC + instructionLength
		} else {
			return cycles, c.PC + instructionLength
		}
	}
}
func LDY(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, pageCrossed := get(c)
		c.Y = v
		c.SetZero(v)
		c.SetSign(v)
		if pageCrossed {
			return cycles + 1, c.PC + instructionLength
		} else {
			return cycles, c.PC + instructionLength
		}
	}
}
func LSR(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, pageCrossed := get(c)
		c.Carry = v&1 != 0
		v >>= 1
		c.SetZero(v)
		c.SetSign(v)
		set(c, v)
		if pageCrossed {
			return cycles + 1, c.PC + instructionLength
		} else {
			return cycles, c.PC + instructionLength
		}
	}
}
func NOP(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) { return cycles, c.PC + instructionLength }
}

func NOPGET(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		_, pageCrossed := get(c)
		if pageCrossed {
			return cycles + 1, c.PC + instructionLength
		} else {
			return cycles, c.PC + instructionLength
		}
	}
}

func ora(c *CPU, v byte) {
	c.A = c.A | v
	c.SetZero(c.A)
	c.SetSign(c.A)
}

func ORA(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, pageCrossed := get(c)

		ora(c, v)

		if pageCrossed {
			return cycles + 1, c.PC + instructionLength
		} else {
			return cycles, c.PC + instructionLength
		}
	}
}
func PHA(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		c.Push(c.A)
		return cycles, c.PC + instructionLength
	}
}
func PHP(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		c.Push(c.Status() | 16)
		return cycles, c.PC + instructionLength
	}
}
func PLA(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		c.A = c.Pop()
		c.SetZero(c.A)
		c.SetSign(c.A)
		return cycles, c.PC + instructionLength
	}
}
func PLP(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		c.SetStatus(c.Pop() & 0xEF)
		return cycles, c.PC + instructionLength
	}
}

func rol(c *CPU, v byte) byte {
	nc := v&0x80 != 0
	v <<= 1
	if c.Carry {
		v |= 1
	}
	c.Carry = nc
	c.SetZero(v)
	c.SetSign(v)
	return v
}

func ROL(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, _ := get(c)
		v = rol(c, v)
		set(c, v)
		return cycles, c.PC + instructionLength
	}
}

func ror(c *CPU, v byte) byte {
	nc := v&1 != 0
	v >>= 1
	if c.Carry {
		v |= 0x80
	}
	c.Carry = nc
	c.SetZero(v)
	c.SetSign(v)
	return v
}

func ROR(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, _ := get(c)
		v = ror(c, v)
		set(c, v)
		return cycles, c.PC + instructionLength
	}
}
func RTI(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		c.SetStatus(c.Pop())
		return cycles, uint16(c.Pop()) | (uint16(c.Pop()) << 8)
	}
}
func RTS(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		val := (uint16(c.Pop()) | (uint16(c.Pop()) << 8)) + 1
		return cycles, val
	}
}

func sbc(c *CPU, v byte) {
	t := uint16(c.A) - uint16(v)
	if !c.Carry {
		t -= 1
	}
	c.Carry = t < 0x100
	bt := byte(t)
	c.Overflow = ((((c.A ^ v) & 0x80) != 0) && ((c.A^bt)&0x80) != 0)
	c.SetZero(bt)
	c.SetSign(bt)
	c.A = bt
}

func SBC(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, pageCrossed := get(c)
		sbc(c, v)
		if pageCrossed {
			return cycles + 1, c.PC + instructionLength
		} else {
			return cycles, c.PC + instructionLength
		}
	}
}
func SEC(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		c.Carry = true
		return cycles, c.PC + instructionLength
	}
}
func SED(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		c.Decimal = true
		return cycles, c.PC + instructionLength
	}
}
func SEI(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		c.Interrupt = true
		return cycles, c.PC + instructionLength
	}
}
func STA(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		set(c, c.A)
		return cycles, c.PC + instructionLength
	}
}
func STX(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		set(c, c.X)
		return cycles, c.PC + instructionLength
	}
}
func STY(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		set(c, c.Y)
		return cycles, c.PC + instructionLength
	}
}
func TAX(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		c.X = c.A
		c.SetZero(c.X)
		c.SetSign(c.X)
		return cycles, c.PC + instructionLength
	}
}
func TAY(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		c.Y = c.A
		c.SetZero(c.Y)
		c.SetSign(c.Y)
		return cycles, c.PC + instructionLength
	}
}
func TSX(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		c.X = c.SP
		c.SetZero(c.X)
		c.SetSign(c.X)
		return cycles, c.PC + instructionLength
	}
}
func TXA(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		c.A = c.X
		c.SetZero(c.A)
		c.SetSign(c.A)
		return cycles, c.PC + instructionLength
	}
}
func TXS(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		c.SP = c.X
		return cycles, c.PC + instructionLength
	}
}
func TYA(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		c.A = c.Y
		c.SetZero(c.A)
		c.SetSign(c.A)
		return cycles, c.PC + instructionLength
	}
}

func DCP(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, pageCrossed := get(c)
		t := uint16(v)
		t--
		c.Carry = t < 0x100
		v = byte(t)
		set(c, v)

		compare(v, pageCrossed, instructionLength, operand, cycles, c.A, c)

		return cycles, c.PC + instructionLength
	}
}

func ISB(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, _ := get(c)
		v++
		set(c, v)
		sbc(c, v)

		return cycles, c.PC + instructionLength
	}
}

func SLO(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, _ := get(c)

		v = asl(c, v)

		set(c, v)

		ora(c, v)
		return cycles, c.PC + instructionLength
	}
}

func RLA(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, _ := get(c)

		v = rol(c, v)

		set(c, v)

		and(c, v)
		return cycles, c.PC + instructionLength
	}
}

func SRE(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, _ := get(c)

		c.Carry = v&1 != 0
		v >>= 1 //ror(c, v)

		set(c, v)

		eor(c, v)
		return cycles, c.PC + instructionLength
	}
}

func RRA(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(c *CPU) (int, uint16) {
		v, _ := get(c)

		v = ror(c, v)

		set(c, v)

		adc(c, v)
		return cycles, c.PC + instructionLength
	}
}
