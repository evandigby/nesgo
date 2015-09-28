package cpu

var cycleExceptions = map[string]map[int]int{
	opASL: map[int]int{
		AddressZeroPage:  5,
		AddressZeroPageX: 6,
		AddressAbsolute:  6,
		AddressAbsoluteX: 7,
	},
	opBRK: map[int]int{
		AddressRelative: 7,
	},
	opDEC: map[int]int{
		AddressZeroPage:  5,
		AddressZeroPageX: 6,
		AddressAbsolute:  6,
		AddressAbsoluteX: 7,
	},
	opINC: map[int]int{
		AddressZeroPage:  5,
		AddressZeroPageX: 6,
		AddressAbsolute:  6,
		AddressAbsoluteX: 7,
	},
	opJMP: map[int]int{
		AddressAddress: 3,
	},
	opJSR: map[int]int{
		AddressAddress: 6,
	},
	opLSR: map[int]int{
		AddressZeroPage:  5,
		AddressZeroPageX: 6,
		AddressAbsolute:  6,
		AddressAbsoluteX: 7,
	},
	opPHA: map[int]int{
		AddressImplied: 3,
	},
	opPHP: map[int]int{
		AddressImplied: 3,
	},
	opPLA: map[int]int{
		AddressImplied: 4,
	},
	opPLP: map[int]int{
		AddressImplied: 4,
	},
	opRTI: map[int]int{
		AddressImplied: 6,
	},
	opRTS: map[int]int{
		AddressImplied: 6,
	},
	opROL: map[int]int{
		AddressZeroPage:  5,
		AddressZeroPageX: 6,
		AddressAbsolute:  6,
		AddressAbsoluteX: 7,
	},
	opROR: map[int]int{
		AddressZeroPage:  5,
		AddressZeroPageX: 6,
		AddressAbsolute:  6,
		AddressAbsoluteX: 7,
	},
	opSTA: map[int]int{
		AddressAbsoluteX: 5,
		AddressAbsoluteY: 5,
		AddressIndirectY: 6,
	},
	opDCPu: map[int]int{
		AddressZeroPage:  5,
		AddressZeroPageX: 6,
		AddressAbsolute:  6,
		AddressAbsoluteX: 7,
		AddressAbsoluteY: 7,
		AddressIndirectX: 8,
		AddressIndirectY: 8,
	},
	opISBu: map[int]int{
		AddressZeroPage:  5,
		AddressZeroPageX: 6,
		AddressAbsolute:  6,
		AddressAbsoluteX: 7,
		AddressAbsoluteY: 7,
		AddressIndirectX: 8,
		AddressIndirectY: 8,
	},
	opSLOu: map[int]int{
		AddressZeroPage:  5,
		AddressZeroPageX: 6,
		AddressAbsolute:  6,
		AddressAbsoluteX: 7,
		AddressAbsoluteY: 7,
		AddressIndirectX: 8,
		AddressIndirectY: 8,
	},
	opRLAu: map[int]int{
		AddressZeroPage:  5,
		AddressZeroPageX: 6,
		AddressAbsolute:  6,
		AddressAbsoluteX: 7,
		AddressAbsoluteY: 7,
		AddressIndirectX: 8,
		AddressIndirectY: 8,
	},
	opSREu: map[int]int{
		AddressZeroPage:  5,
		AddressZeroPageX: 6,
		AddressAbsolute:  6,
		AddressAbsoluteX: 7,
		AddressAbsoluteY: 7,
		AddressIndirectX: 8,
		AddressIndirectY: 8,
	},
	opRRAu: map[int]int{
		AddressZeroPage:  5,
		AddressZeroPageX: 6,
		AddressAbsolute:  6,
		AddressAbsoluteX: 7,
		AddressAbsoluteY: 7,
		AddressIndirectX: 8,
		AddressIndirectY: 8,
	},
}

func getCycles(instruction string, addressMode int) int {
	var cycles int

	switch addressMode {
	case AddressAccumulator:
		return 2
	case AddressImmediate:
		return 2
	case AddressZeroPage:
		cycles = 3 // With Exceptions
	case AddressZeroPageX:
		cycles = 4 // With Exceptions
	case AddressZeroPageY:
		return 4
	case AddressAbsolute:
		cycles = 4 // With Exceptions
	case AddressAbsoluteX:
		cycles = 4
	case AddressAbsoluteY:
		cycles = 4 // With Exceptions
	case AddressImplied:
		cycles = 2 // With Exceptions
	case AddressRelative:
		cycles = 2 // With Exceptions
	case AddressIndirectX:
		cycles = 6 // With Exceptions
	case AddressIndirectY:
		cycles = 5 // With Exceptions
	case AddressIndirect:
		return 5
	}
	// AddressAddress Only in exceptions (2 cases)

	if m, ok := cycleExceptions[instruction]; ok {
		if c, ok := m[addressMode]; ok {
			return c
		}
	}

	return cycles
}
