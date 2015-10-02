package ppu

import (
	"image"

	"github.com/evandigby/nesgo/nes"
	"github.com/evandigby/nesgo/rom"
)

type Registers struct {
	PPUCTRL   byte
	PPUMASK   byte
	PPUSTATUS byte
	OAMADDR   byte
	//OAMDATA   byte
	PPUSCROLL byte
	PPUADDR   byte
	PPUDATA   byte
	OAMDMA    byte
}

type PPU struct {
	sync   chan int
	Memory []*byte
	OAM    []*byte

	vramAddr  uint16
	tvramAddr uint16

	renderer Renderer
	rom      rom.ROM

	cycle        int
	scanLine     int
	odd          bool
	nmi          bool
	rendering    bool
	scrollToggle bool
	scrollX      byte
	scrollY      byte
	addrToggle   bool
	addr         uint16

	*Registers

	MemoryMap nes.MemoryMap

	frame *image.NRGBA
}

func NewPPU(sync chan int, rom rom.ROM, renderer Renderer) *PPU {
	tm := make([]byte, 0x4000)
	m := make([]*byte, 0x4000)

	for i := range m {
		m[i] = &tm[i]
	}

	chrRom := rom.CharRom()
	for i := range chrRom {
		m[i] = chrRom[i]
	}

	// Mirror some Memory
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

	ppu := &PPU{
		sync:      sync,
		Memory:    m,
		OAM:       oam,
		renderer:  renderer,
		rom:       rom,
		Registers: &Registers{},
	}

	mm := nes.MemoryMap{
		0x2000: &MappedRegister{func(debug bool) byte { return ppu.PPUCTRL }, func(val byte) { ppu.PPUCTRL = val }},
		0x2001: &MappedRegister{func(debug bool) byte { return ppu.PPUMASK }, func(val byte) { ppu.PPUMASK = val }},
		0x2002: &MappedRegister{ppu.ReadPPUStatus, func(val byte) { ppu.PPUSTATUS = val }},
		0x2003: &MappedRegister{func(debug bool) byte { return ppu.OAMADDR }, func(val byte) { ppu.OAMADDR = val }},
		0x2004: &MappedRegister{ppu.ReadOAMDATA, ppu.WriteOAMDATA},
		0x2005: &MappedRegister{func(debug bool) byte { return ppu.PPUSCROLL }, ppu.WritePPUSCROLL},
		0x2006: &MappedRegister{func(debug bool) byte { return ppu.PPUADDR }, ppu.WritePPUADDR},
		0x2007: &MappedRegister{ppu.ReadPPUDATA, ppu.WritePPUDATA},
		0x4014: &MappedRegister{func(debug bool) byte {
			return ppu.OAMDMA
		},
			func(val byte) {
				ppu.OAMDMA = val
			},
		},
	}

	ppu.MemoryMap = mm

	return ppu
}

func (p *PPU) ReadPPUStatus(debug bool) byte {
	val := p.PPUSTATUS

	if debug {
		return val
	}
	p.PPUSTATUS &= 0x7F
	p.scrollToggle = false

	return val
}

func (p *PPU) WriteOAMDATA(value byte) {
	if !p.rendering {
		*p.OAM[p.OAMADDR] = value
		p.OAMADDR++
	}
}

func (p *PPU) ReadOAMDATA(debug bool) byte {
	return *p.OAM[p.OAMADDR]
}

func (p *PPU) WritePPUSCROLL(value byte) {
	if !p.scrollToggle {
		p.scrollX = value
	} else {
		p.scrollY = value
	}

	p.scrollToggle = !p.scrollToggle
}

func (p *PPU) WritePPUADDR(value byte) {
	if !p.addrToggle {
		p.addr &= 0x00FF
		p.addr |= uint16(value) << 8
	} else {
		p.addr &= 0xFF00
		p.addr |= uint16(value)
	}

	p.addrToggle = !p.addrToggle
}

func (p *PPU) WritePPUDATA(value byte) {
	*p.Memory[p.addr] = value
	if p.PPUCTRL&4 > 0 {
		p.addr += 32
	} else {
		p.addr++
	}
}

func (p *PPU) ReadPPUDATA(debug bool) byte {
	val := *p.Memory[p.addr]

	if debug {
		return val
	}

	if p.PPUCTRL&4 > 0 {
		p.addr += 32
	} else {
		p.addr++
	}

	return val
}

type MappedRegister struct {
	reader func(debug bool) byte
	writer func(val byte)
}

func (r *MappedRegister) Read(debug bool) byte {
	return r.reader(debug)
}

func (r *MappedRegister) Write(val byte) {
	r.writer(val)
}

func (p *PPU) PowerOn() {
	p.PPUCTRL = 0x00
	p.PPUMASK = 0x00
	p.PPUSTATUS = 0xA0
	p.OAMADDR = 0x00
	p.OAMADDR = 0x00
	p.PPUSCROLL = 0x00
	p.PPUADDR = 0x00
	p.PPUDATA = 0x00

	p.odd = false
}

func (p *PPU) Reset() {
	p.PPUCTRL = 0x00
	p.PPUMASK = 0x00
	p.PPUSCROLL = 0x00
	p.PPUDATA = 0x00

	p.odd = false
}

func (p *PPU) visible() {
	if p.cycle == 0 {

	} else if p.cycle > 0 && p.cycle <= 256 {

	} else if p.cycle > 256 && p.cycle <= 320 {
		p.OAMADDR = 0x00
	} else if p.cycle > 320 && p.cycle <= 336 {

	} else if p.cycle > 336 && p.cycle <= 340 {

	}
}

func (p *PPU) vBlank() {
	if p.cycle == 1 {
		p.PPUSTATUS |= 0x80
	}
}

func (p *PPU) preRender() {
	if p.cycle == 1 {
		p.PPUSTATUS &= 0x7F
		p.frame = image.NewNRGBA(image.Rect(0, 0, 256, 240))
	} else if p.cycle > 256 && p.cycle <= 320 {
		p.OAMADDR = 0x00
	}
}

func (p *PPU) postRender() {
	p.renderer.Render(p.frame)
}

func (p *PPU) runCycle() {
	if p.scanLine == 261 {
		p.rendering = true
		p.preRender()
	} else if p.scanLine > 0 && p.scanLine <= 239 {
		p.rendering = true
		p.visible()
	} else if p.scanLine == 240 {
		p.rendering = false
		p.postRender()
	} else if p.scanLine > 240 && p.scanLine <= 260 {
		p.rendering = false
		p.vBlank()
	}
}

func (p *PPU) execute() {
	p.PowerOn()

	for {
		<-p.sync
		if p.cycle > 340 {
			p.cycle = 0
			p.scanLine++

			if p.scanLine > 261 {
				p.scanLine = -1
			}
		}

		//fmt.Printf("Running Cycle %v on Scan Line %v\n", p.cycle, p.scanLine)

		p.runCycle()

		p.sync <- p.cycle
		p.cycle++
	}
}

func (p *PPU) Run() {
	go p.execute()
}
