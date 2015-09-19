package cpu

func getInstruction(op byte) string {
	switch op {
	case 0x69, 0x65, 0x75, 0x6D, 0x7D, 0x79, 0x61, 0x71:
		return opADC
	case 0x29, 0x25, 0x35, 0x2D, 0x3D, 0x39, 0x21, 0x31:
		return opAND
	case 0x0A, 0x06, 0x16, 0x0E, 0x1E:
		return opASL
	case 0x90:
		return opBCC
	case 0xB0:
		return opBCS
	case 0xF0:
		return opBEQ
	case 0x24, 0x2C:
		return opBIT
	case 0x30:
		return opBMI
	case 0xD0:
		return opBNE
	case 0x10:
		return opBPL
	case 0x00:
		return opBRK
	case 0x50:
		return opBVC
	case 0x70:
		return opBVS
	case 0x18:
		return opCLC
	case 0xD8:
		return opCLD
	case 0x58:
		return opCLI
	case 0xB8:
		return opCLV
	case 0xC9, 0xC5, 0xD5, 0xCD, 0xDD, 0xD9, 0xC1, 0xD1:
		return opCMP
	case 0xE0, 0xE4, 0xEC:
		return opCPX
	case 0xC0, 0xC4, 0xCC:
		return opCPY
	case 0xC6, 0xD6, 0xCE, 0xDE:
		return opDEC
	case 0xCA:
		return opDEX
	case 0x88:
		return opDEY
	case 0x49, 0x45, 0x55, 0x4D, 0x5D, 0x59, 0x41, 0x51:
		return opEOR
	case 0xE6, 0xF6, 0xEE, 0xFE:
		return opINC
	case 0xE8:
		return opINX
	case 0xC8:
		return opINY
	case 0x4C, 0x6C:
		return opJMP
	case 0x20:
		return opJSR
	case 0xA9, 0xA5, 0xB5, 0xAD, 0xBD, 0xB9, 0xA1, 0xB1:
		return opLDA
	case 0xA2, 0xA6, 0xB6, 0xAE, 0xBE:
		return opLDX
	case 0xA0, 0xA4, 0xB4, 0xAC, 0xBC:
		return opLDY
	case 0x4A, 0x46, 0x56, 0x4E, 0x5E:
		return opLSR
	case 0xEA:
		return opNOP
	case 0x09, 0x05, 0x15, 0x0D, 0x1D, 0x19, 0x01, 0x11:
		return opORA
	case 0x48:
		return opPHA
	case 0x08:
		return opPHP
	case 0x68:
		return opPLA
	case 0x28:
		return opPLP
	case 0x2A, 0x26, 0x36, 0x2E, 0x3E:
		return opROL
	case 0x6A, 0x66, 0x76, 0x6E, 0x7E:
		return opROR
	case 0x40:
		return opRTI
	case 0x60:
		return opRTS
	case 0xE9, 0xE5, 0xF5, 0xED, 0xFD, 0xF9, 0xE1, 0xF1:
		return opSBC
	case 0x38:
		return opSEC
	case 0xF8:
		return opSED
	case 0x78:
		return opSEI
	case 0x85, 0x95, 0x8D, 0x9D, 0x99, 0x81, 0x91:
		return opSTA
	case 0x86, 0x96, 0x8E:
		return opSTX
	case 0x84, 0x94, 0x8C:
		return opSTY
	case 0xAA:
		return opTAX
	case 0xA8:
		return opTAY
	case 0xBA:
		return opTSX
	case 0x8A:
		return opTXA
	case 0x9A:
		return opTXS
	case 0x98:
		return opTYA
	case 0x04, 0x44, 0x64, 0x0C, 0x14, 0x34, 0x54, 0x74, 0xD4, 0xF4, 0x1A, 0x3A, 0x5A, 0x7A, 0xDA, 0xFA, 0x80,
		0x1C, 0x3C, 0x5C, 0x7C, 0xDC, 0xFC:
		return opNOPu
	case 0xA7, 0xB7, 0xAF, 0xBF, 0xA3, 0xB3:
		return opLAXu
	default:
		return opUNK
	}
}
