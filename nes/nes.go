package nes

import "github.com/evandigby/nesgo/rom"

type ByteReadWriter interface {
	Read(debug bool) byte
	Write(val byte)
}

type MemoryMap map[uint16]ByteReadWriter

type NES struct {
	Memory    []*byte `json:"-"`
	Stack     []*byte `json:"-"`
	Cartridge []*byte `json:"-"`

	memoryMap map[uint16]ByteReadWriter

	Debug bool
}

const MemSize = 0xFFFF

func NewNES(memoryMap MemoryMap) *NES {
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

	// Cartridge Memory helper
	c := m[0x8000 : MemSize+1]

	return &NES{m, s, c, memoryMap, false}
}

func (n *NES) LoadRom(r rom.ROM) {
	mirror := r.Pages() == 1
	for i, v := range r.ProgramRom() {
		*n.Cartridge[i] = *v
		if mirror {
			*n.Cartridge[i+0x4000] = *v
		}
	}
}

func (n *NES) PowerUp() {
	/*
		for i := 0; i < 0x0800; i++ {
			*n.Memory[i] = 0xFF
		}
	*/
	*n.Memory[0x0008] = 0xF7
	*n.Memory[0x0009] = 0xEF
	*n.Memory[0x000A] = 0xDF
	*n.Memory[0x000F] = 0xBF
	*n.Memory[0x4017] = 0x00
	*n.Memory[0x4015] = 0x00
	for i := 0x4000; i < 0x4010; i++ {
		*n.Memory[i] = 0x00
	}
}

func (n *NES) Reset() {
	*n.Memory[0x4015] = 0x00
}

func (n *NES) Get(address uint16) byte {
	if m, ok := n.memoryMap[address]; ok {
		return m.Read(n.Debug)
	}

	return *n.Memory[address]
}

func (n *NES) Set(address uint16, value byte) {
	if m, ok := n.memoryMap[address]; ok {
		m.Write(value)
	}

	*n.Memory[address] = value
}
