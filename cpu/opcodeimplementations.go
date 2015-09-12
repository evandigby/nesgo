// Decimal Mode Not Implemented -- doesn't exist in NES
package cpu

func ADC(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, c := s.GetValue(addressMode, operand)
		t := uint16(v) + uint16(s.A)
		if s.Carry {
			t += 1
		}
		s.Carry = t > 0xFF
		bt := byte(t)
		s.Overflow = (s.A^v)&(v^bt)&0x80 != 0
		s.SetZero(bt)
		s.SetSign(bt)
		s.A = bt
		return cycles + c, s.PC + instructionLength
	}
}
func AND(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, c := s.GetValue(addressMode, operand)
		v = v & s.A
		s.SetSign(v)
		s.SetZero(v)
		s.A = byte(v)
		return cycles + c, s.PC + instructionLength
	}
}
func ASL(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, c := s.GetValue(addressMode, operand)

		s.Carry = v&0x80 != 0
		v <<= 1
		v &= 0xFF
		s.SetSign(v)
		s.SetZero(v)

		return cycles + c, s.PC + instructionLength
	}
}

func branch(b bool, s *State, addressMode int, instructionLength uint16, operand uint16, cycles int) (int, uint16) {
	if b {
		v, c := s.CalculateAddress(addressMode, operand)

		return cycles + c + 1, v
	} else {
		return cycles, s.PC + instructionLength
	}
}

func BCC(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		return branch(!s.Carry, s, addressMode, instructionLength, operand, cycles)
	}
}
func BCS(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		return branch(s.Carry, s, addressMode, instructionLength, operand, cycles)
	}
}
func BEQ(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		return branch(!s.Zero, s, addressMode, instructionLength, operand, cycles)
	}
}
func BIT(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, c := s.GetValue(addressMode, operand)

		r := s.A & byte(v)
		s.SetZero(r)
		s.Overflow = r&0x40 != 0
		s.Sign = r&0x80 != 0

		return cycles + c, s.PC + instructionLength
	}
}
func BMI(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		return branch(s.Sign, s, addressMode, instructionLength, operand, cycles)
	}
}

func BNE(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		return branch(!s.Zero, s, addressMode, instructionLength, operand, cycles)
	}
}
func BPL(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		return branch(!s.Sign, s, addressMode, instructionLength, operand, cycles)
	}
}
func BRK(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.PC++
		s.Push(byte(s.PC >> 8))
		s.Push(byte(s.PC))
		s.Break = true
		s.Push(s.Status())
		s.Interrupt = true
		lsd, _ := s.GetValue(AddressAbsolute, 0xFFFE)
		msd, _ := s.GetValue(AddressAbsolute, 0xFFFE)
		return cycles, uint16(lsd) | (uint16(msd) << 8)
	}
}
func BVC(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		return branch(!s.Overflow, s, addressMode, instructionLength, operand, cycles)
	}
}
func BVS(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		return branch(s.Overflow, s, addressMode, instructionLength, operand, cycles)
	}
}
func CLC(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Carry = false
		return cycles, s.PC + instructionLength
	}
}
func CLD(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Decimal = false
		return cycles, s.PC + instructionLength
	}
}
func CLI(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Interrupt = false
		return cycles, s.PC + instructionLength
	}
}
func CLV(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Overflow = false
		return cycles, s.PC + instructionLength
	}
}

func compare(addressMode int, instructionLength uint16, operand uint16, cycles int, V byte, s *State) (int, uint16) {
	v, c := s.GetValue(addressMode, operand)
	t := V - v

	s.Carry = V >= v
	s.Zero = V == v
	s.SetSign(t)
	return cycles + c, s.PC + instructionLength
}

func CMP(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		return compare(addressMode, instructionLength, operand, cycles, s.A, s)
	}
}
func CPX(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		return compare(addressMode, instructionLength, operand, cycles, s.X, s)
	}
}
func CPY(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		return compare(addressMode, instructionLength, operand, cycles, s.Y, s)
	}
}
func DEC(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, _ := s.GetValue(addressMode, operand)
		v -= 1
		s.SetZero(v)
		s.SetSign(v)
		s.SetValue(addressMode, operand, v)
		return cycles, s.PC + instructionLength
	}
}
func DEX(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.X -= 1
		s.SetZero(s.X)
		s.SetSign(s.X)
		return cycles, s.PC + instructionLength
	}
}
func DEY(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Y -= 1
		s.SetZero(s.Y)
		s.SetSign(s.Y)
		return cycles, s.PC + instructionLength
	}
}
func EOR(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, c := s.GetValue(addressMode, operand)
		s.A = s.A ^ v
		s.SetZero(s.A)
		s.SetSign(s.A)
		return cycles + c, s.PC + instructionLength
	}
}
func INC(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, _ := s.GetValue(addressMode, operand)
		v += 1
		s.SetZero(v)
		s.SetSign(v)
		s.SetValue(addressMode, operand, v)
		return cycles, s.PC + instructionLength
	}
}
func INX(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.X += 1
		s.SetZero(s.X)
		s.SetSign(s.X)
		return cycles, s.PC + instructionLength
	}
}
func INY(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Y += 1
		s.SetZero(s.Y)
		s.SetSign(s.Y)
		return cycles, s.PC + instructionLength
	}
}
func JMP(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, _ := s.CalculateAddress(addressMode, operand)
		return cycles, v
	}
}
func JSR(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v := s.PC - 1
		s.Push(byte(v >> 8))
		s.Push(byte(v))
		return cycles, operand
	}
}
func LDA(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, c := s.GetValue(addressMode, operand)
		s.A = v
		s.SetZero(v)
		s.SetSign(v)
		return cycles + c, s.PC + instructionLength
	}
}
func LDX(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, c := s.GetValue(addressMode, operand)
		s.X = v
		s.SetZero(v)
		s.SetSign(v)
		return cycles + c, s.PC + instructionLength
	}
}
func LDY(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, c := s.GetValue(addressMode, operand)
		s.Y = v
		s.SetZero(v)
		s.SetSign(v)
		return cycles + c, s.PC + instructionLength
	}
}
func LSR(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, c := s.GetValue(addressMode, operand)
		s.Carry = v&1 != 0
		v >>= 1
		s.SetZero(v)
		s.SetSign(v)
		s.SetValue(addressMode, operand, v)
		return cycles + c, s.PC + instructionLength
	}
}
func NOP(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) { return cycles, s.PC + instructionLength }
}
func ORA(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, c := s.GetValue(addressMode, operand)
		s.A = s.A | v
		s.SetZero(s.A)
		s.SetSign(s.A)
		return cycles + c, s.PC + instructionLength
	}
}
func PHA(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Push(s.A)
		return cycles, s.PC + instructionLength
	}
}
func PHP(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Push(s.Status())
		return cycles, s.PC + instructionLength
	}
}
func PLA(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.A = s.Pop()
		return cycles, s.PC + instructionLength
	}
}
func PLP(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.SetStatus(s.Pop())
		return cycles, s.PC + instructionLength
	}
}
func ROL(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, _ := s.GetValue(addressMode, operand)
		nc := v&0x80 != 0
		v <<= 1
		if s.Carry {
			v |= 1
		}
		s.Carry = nc
		s.SetZero(v)
		s.SetSign(v)
		s.SetValue(addressMode, operand, v)
		return cycles, s.PC + instructionLength
	}
}
func ROR(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, _ := s.GetValue(addressMode, operand)
		nc := v&1 != 0
		v >>= 1
		if s.Carry {
			v |= 0x80
		}
		s.Carry = nc
		s.SetZero(v)
		s.SetSign(v)
		s.SetValue(addressMode, operand, v)
		return cycles, s.PC + instructionLength
	}
}
func RTI(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.SetStatus(s.Pop())
		return cycles, uint16(s.Pop()) | (uint16(s.Pop()) << 8)
	}
}
func RTS(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		return cycles, uint16(s.Pop()) | (uint16(s.Pop()) << 8)
	}
}
func SBC(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		v, c := s.GetValue(addressMode, operand)
		t := uint16(s.A) - uint16(v)
		if !s.Carry {
			t -= 1
		}
		s.Carry = t < 0x100
		bt := byte(t)
		s.Overflow = (s.A^v)&(v^bt)&0x80 == 0
		s.SetZero(bt)
		s.SetSign(bt)
		s.A = bt
		return cycles + c, s.PC + instructionLength
	}
}
func SEC(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Carry = true
		return cycles, s.PC + instructionLength
	}
}
func SED(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Decimal = true
		return cycles, s.PC + instructionLength
	}
}
func SEI(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Interrupt = true
		return cycles, s.PC + instructionLength
	}
}
func STA(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.SetValue(addressMode, operand, s.A)
		return cycles, s.PC + instructionLength
	}
}
func STX(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.SetValue(addressMode, operand, s.X)
		return cycles, s.PC + instructionLength
	}
}
func STY(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.SetValue(addressMode, operand, s.Y)
		return cycles, s.PC + instructionLength
	}
}
func TAX(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.X = s.A
		s.SetZero(s.X)
		s.SetSign(s.X)
		return cycles, s.PC + instructionLength
	}
}
func TAY(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.Y = s.A
		s.SetZero(s.Y)
		s.SetSign(s.Y)
		return cycles, s.PC + instructionLength
	}
}
func TSX(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.X = s.SP
		s.SetZero(s.X)
		s.SetSign(s.X)
		return cycles, s.PC + instructionLength
	}
}
func TXA(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.A = s.X
		s.SetZero(s.A)
		s.SetSign(s.A)
		return cycles, s.PC + instructionLength
	}
}
func TXS(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.SP = s.X
		s.SetZero(s.SP)
		s.SetSign(s.SP)
		return cycles, s.PC + instructionLength
	}
}
func TYA(addressMode int, instructionLength uint16, operand uint16, cycles int) Executer {
	return func(s *State) (int, uint16) {
		s.A = s.Y
		s.SetZero(s.A)
		s.SetSign(s.A)
		return cycles, s.PC + instructionLength
	}
}
