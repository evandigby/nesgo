package ppu

type PPU struct {
	sync   chan int
	memory []*byte
	oam    []*byte

	vramAddr  uint16
	tvramAddr uint16
}

/*
func NewPPU(sync chan int) *PPU {
	tm := make([]byte, 0x4000)
	m := make([]*byte, 0x4000)

	for i := range m {
		m[i] = &tm[i]
	}

	// Mirror some memory
	for i := 0x2000; i < 0x2F00; i++ {
		m[i+0x1000] = m[i]
	}
	for i := 0x3F00; i < 0x3F20; i++ {
		m[i+0x0020] = m[i]
	}

	toam := make([]byte, 0x100)
	oam := make([]*byte, 0x100)

	for i := range oam {
		oam[i] = &toam[i]
	}

	return &PPU{sync, m, oam}
}
*/
