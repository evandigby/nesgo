package nes

type ByteReadWriter interface {
	Read(debug bool) byte
	Write(val byte)
}

type MemoryMap map[uint16]ByteReadWriter
