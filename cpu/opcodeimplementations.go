// Decimal Mode Not Implemented -- doesn't exist in NES
package cpu

func ADC(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, c := get(s)
		if c {
			cycles++
		}
		t := uint16(v) + uint16(s.A)
		if s.Carry {
			t += 1
		}
		bt := byte(t)
		s.Carry = t > 0xFF
		s.Overflow = ((((s.A ^ v) & 0x80) == 0) && ((s.A^bt)&0x80) != 0)
		s.SetZero(bt)
		s.SetSign(bt)
		s.A = bt
		return cycles, s.PC + instructionLength
	}
}
func AND(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, c := get(s)
		if c {
			cycles++
		}
		v = v & s.A
		s.SetSign(v)
		s.SetZero(v)
		s.A = byte(v)
		return cycles, s.PC + instructionLength
	}
}
func ASL(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, c := get(s)
		if c {
			cycles++
		}

		s.Carry = v&0x80 != 0
		v <<= 1
		v &= 0xFF
		s.SetSign(v)
		s.SetZero(v)
		set(s, v)
		return cycles, s.PC + instructionLength
	}
}

func branch(b bool, s *State, v uint16, c bool, instructionLength uint16, operand uint16, cycles int) (int, uint16) {
	if b {
		if c {
			cycles += 2
		} else {
			cycles++
		}

		return cycles, v
	} else {
		return cycles, s.PC + instructionLength
	}
}

func BCC(address, instructionLength, operand uint16, cycles int) Executer {
	v, c := calculateRelativeAddress(instructionLength, operand, address)
	return func(s *State) (int, uint16) {
		return branch(!s.Carry, s, v, c, instructionLength, operand, cycles)
	}
}
func BCS(address, instructionLength, operand uint16, cycles int) Executer {
	v, c := calculateRelativeAddress(instructionLength, operand, address)
	return func(s *State) (int, uint16) {
		return branch(s.Carry, s, v, c, instructionLength, operand, cycles)
	}
}
func BEQ(address, instructionLength, operand uint16, cycles int) Executer {
	v, c := calculateRelativeAddress(instructionLength, operand, address)
	return func(s *State) (int, uint16) {
		return branch(s.Zero, s, v, c, instructionLength, operand, cycles)
	}
}
func BIT(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, c := get(s)
		if c {
			cycles++
		}

		r := s.A & byte(v)
		s.SetZero(r)
		s.Overflow = byte(v)&0x40 != 0
		s.SetSign(byte(v))

		return cycles, s.PC + instructionLength
	}
}
func BMI(address, instructionLength, operand uint16, cycles int) Executer {
	v, c := calculateRelativeAddress(instructionLength, operand, address)
	return func(s *State) (int, uint16) {
		return branch(s.Negative, s, v, c, instructionLength, operand, cycles)
	}
}

func BNE(address, instructionLength, operand uint16, cycles int) Executer {
	v, c := calculateRelativeAddress(instructionLength, operand, address)
	return func(s *State) (int, uint16) {
		return branch(!s.Zero, s, v, c, instructionLength, operand, cycles)
	}
}
func BPL(address, instructionLength, operand uint16, cycles int) Executer {
	v, c := calculateRelativeAddress(instructionLength, operand, address)
	return func(s *State) (int, uint16) {
		return branch(!s.Negative, s, v, c, instructionLength, operand, cycles)
	}
}
func BRK(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.PC++
		s.Push(byte(s.PC >> 8))
		s.Push(byte(s.PC))
		s.Break = true
		s.Push(s.Status())
		s.Interrupt = true
		lsd := *s.Memory[0xFFFE]
		msd := *s.Memory[0xFFFF]
		return cycles, uint16(lsd) | (uint16(msd) << 8)
	}
}
func BVC(address, instructionLength, operand uint16, cycles int) Executer {
	v, c := calculateRelativeAddress(instructionLength, operand, address)
	return func(s *State) (int, uint16) {
		return branch(!s.Overflow, s, v, c, instructionLength, operand, cycles)
	}
}
func BVS(address, instructionLength, operand uint16, cycles int) Executer {
	v, c := calculateRelativeAddress(instructionLength, operand, address)
	return func(s *State) (int, uint16) {
		return branch(s.Overflow, s, v, c, instructionLength, operand, cycles)
	}
}
func CLC(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Carry = false
		return cycles, s.PC + instructionLength
	}
}
func CLD(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Decimal = false
		return cycles, s.PC + instructionLength
	}
}
func CLI(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Interrupt = false
		return cycles, s.PC + instructionLength
	}
}
func CLV(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Overflow = false
		return cycles, s.PC + instructionLength
	}
}

func compare(get Getter, instructionLength uint16, operand uint16, cycles int, V byte, s *State) (int, uint16) {
	v, c := get(s)
	if c {
		cycles++
	}
	t := V - v

	s.Carry = V >= v
	s.Zero = V == v
	s.SetSign(t)
	return cycles, s.PC + instructionLength
}

func CMP(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		return compare(get, instructionLength, operand, cycles, s.A, s)
	}
}
func CPX(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		return compare(get, instructionLength, operand, cycles, s.X, s)
	}
}
func CPY(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		return compare(get, instructionLength, operand, cycles, s.Y, s)
	}
}
func DEC(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, _ := get(s)
		v -= 1
		s.SetZero(v)
		s.SetSign(v)
		set(s, v)
		return cycles, s.PC + instructionLength
	}
}
func DEX(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.X -= 1
		s.SetZero(s.X)
		s.SetSign(s.X)
		return cycles, s.PC + instructionLength
	}
}
func DEY(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Y -= 1
		s.SetZero(s.Y)
		s.SetSign(s.Y)
		return cycles, s.PC + instructionLength
	}
}
func EOR(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, c := get(s)
		if c {
			cycles++
		}
		s.A = s.A ^ v
		s.SetZero(s.A)
		s.SetSign(s.A)
		return cycles, s.PC + instructionLength
	}
}
func INC(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, _ := get(s)
		v += 1
		s.SetZero(v)
		s.SetSign(v)
		set(s, v)
		return cycles, s.PC + instructionLength
	}
}
func INX(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.X += 1
		s.SetZero(s.X)
		s.SetSign(s.X)
		return cycles, s.PC + instructionLength
	}
}
func INY(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Y += 1
		s.SetZero(s.Y)
		s.SetSign(s.Y)
		return cycles, s.PC + instructionLength
	}
}
func JMP(get AddressGetter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		a, _ := get(s)
		return cycles, a
	}
}
func JSR(get AddressGetter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v := s.PC + instructionLength - 1
		s.Push(byte(v >> 8))
		s.Push(byte(v))
		a, _ := get(s)
		return cycles, a
	}
}
func LDA(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, c := get(s)
		if c {
			cycles++
		}
		s.A = v
		s.SetZero(v)
		s.SetSign(v)
		return cycles, s.PC + instructionLength
	}
}
func LAX(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, c := get(s)
		if c {
			cycles++
		}
		s.A = v
		s.X = v
		s.SetZero(v)
		s.SetSign(v)
		return cycles, s.PC + instructionLength
	}
}
func LDX(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, c := get(s)
		if c {
			cycles++
		}
		s.X = v
		s.SetZero(v)
		s.SetSign(v)
		return cycles, s.PC + instructionLength
	}
}
func LDY(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, c := get(s)
		if c {
			cycles++
		}
		s.Y = v
		s.SetZero(v)
		s.SetSign(v)
		return cycles, s.PC + instructionLength
	}
}
func LSR(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, c := get(s)
		if c {
			cycles++
		}
		s.Carry = v&1 != 0
		v >>= 1
		s.SetZero(v)
		s.SetSign(v)
		set(s, v)
		return cycles, s.PC + instructionLength
	}
}
func NOP(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) { return cycles, s.PC + instructionLength }
}

func NOPGET(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		_, c := get(s)
		rc := cycles
		if c {
			rc++
		}
		return rc, s.PC + instructionLength
	}
}
func ORA(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, c := get(s)
		if c {
			cycles++
		}
		s.A = s.A | v
		s.SetZero(s.A)
		s.SetSign(s.A)
		return cycles, s.PC + instructionLength
	}
}
func PHA(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Push(s.A)
		return cycles, s.PC + instructionLength
	}
}
func PHP(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Push(s.Status() | 16)
		return cycles, s.PC + instructionLength
	}
}
func PLA(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.A = s.Pop()
		s.SetZero(s.A)
		s.SetSign(s.A)
		return cycles, s.PC + instructionLength
	}
}
func PLP(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.SetStatus(s.Pop() & 0xEF)
		return cycles, s.PC + instructionLength
	}
}
func ROL(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, _ := get(s)
		nc := v&0x80 != 0
		v <<= 1
		if s.Carry {
			v |= 1
		}
		s.Carry = nc
		s.SetZero(v)
		s.SetSign(v)
		set(s, v)
		return cycles, s.PC + instructionLength
	}
}
func ROR(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, _ := get(s)
		nc := v&1 != 0
		v >>= 1
		if s.Carry {
			v |= 0x80
		}
		s.Carry = nc
		s.SetZero(v)
		s.SetSign(v)
		set(s, v)
		return cycles, s.PC + instructionLength
	}
}
func RTI(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.SetStatus(s.Pop())
		return cycles, uint16(s.Pop()) | (uint16(s.Pop()) << 8)
	}
}
func RTS(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		val := (uint16(s.Pop()) | (uint16(s.Pop()) << 8)) + 1
		return cycles, val
	}
}
func SBC(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, c := get(s)
		if c {
			cycles++
		}
		t := uint16(s.A) - uint16(v)
		if !s.Carry {
			t -= 1
		}
		s.Carry = t < 0x100
		bt := byte(t)
		s.Overflow = ((((s.A ^ v) & 0x80) != 0) && ((s.A^bt)&0x80) != 0)
		s.SetZero(bt)
		s.SetSign(bt)
		s.A = bt
		return cycles, s.PC + instructionLength
	}
}
func SEC(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Carry = true
		return cycles, s.PC + instructionLength
	}
}
func SED(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Decimal = true
		return cycles, s.PC + instructionLength
	}
}
func SEI(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Interrupt = true
		return cycles, s.PC + instructionLength
	}
}
func STA(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		set(s, s.A)
		return cycles, s.PC + instructionLength
	}
}
func STX(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		set(s, s.X)
		return cycles, s.PC + instructionLength
	}
}
func STY(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		set(s, s.Y)
		return cycles, s.PC + instructionLength
	}
}
func TAX(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.X = s.A
		s.SetZero(s.X)
		s.SetSign(s.X)
		return cycles, s.PC + instructionLength
	}
}
func TAY(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Y = s.A
		s.SetZero(s.Y)
		s.SetSign(s.Y)
		return cycles, s.PC + instructionLength
	}
}
func TSX(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.X = s.SP
		s.SetZero(s.X)
		s.SetSign(s.X)
		return cycles, s.PC + instructionLength
	}
}
func TXA(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.A = s.X
		s.SetZero(s.A)
		s.SetSign(s.A)
		return cycles, s.PC + instructionLength
	}
}
func TXS(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.SP = s.X
		return cycles, s.PC + instructionLength
	}
}
func TYA(get Getter, set Setter, address, instructionLength, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.A = s.Y
		s.SetZero(s.A)
		s.SetSign(s.A)
		return cycles, s.PC + instructionLength
	}
}
