package rom

import (
	"errors"
	"io"
	"io/ioutil"
)

type INES struct {
	raw        []byte
	header     []byte
	trainer    []byte
	programRom []byte
	charRom    []byte
	instRom    []byte
	pRom       []byte

	fourScreen bool
	hasTrainer bool
	sram       bool
	hMirroring bool
	vMirroring bool

	ines2        bool
	playChoice10 bool
	vsUnisystem  bool

	mapper uint8
}

func (r *INES) Trainer() []byte           { return r.trainer }
func (r *INES) ProgramRom() []byte        { return r.programRom }
func (r *INES) CharRom() []byte           { return r.charRom }
func (r *INES) PlayChoiceInstRom() []byte { return r.instRom }
func (r *INES) PlayChoicePRom() []byte    { return r.pRom }

const (
	headerSize         int = 16
	trainerSize            = 512
	programRomPageSize     = 16384
	charRomPageSize        = 8192
	playChoicesize         = 8192
	pRomDataSize           = 16
	pRomCounterOutSize     = 16
)

func NewINES(reader io.Reader) (ROM, error) {
	raw, err := ioutil.ReadAll(reader)

	if err != nil {
		return &INES{}, errors.New("Unable to read iNES format from stream")
	}

	r := &INES{raw: raw}

	r.header = raw[0:headerSize]

	//$4E $45 $53 $1A (NES + EOF)
	if r.header[0] != 0x4E || r.header[1] != 0x45 || r.header[2] != 0x53 || r.header[3] != 0x1A {
		return r, errors.New("Not a valid iNES format")
	}

	flags6 := uint8(r.header[6])
	flags7 := uint8(r.header[7])

	r.fourScreen = flags6&(1<<3) != 0
	r.hasTrainer = flags6&(1<<2) != 0
	r.sram = flags6&(1<<1) != 0
	r.vMirroring = flags6&1 != 0
	r.hMirroring = !r.vMirroring

	r.mapper = (flags7 & 0xF0) | (flags6 >> 4)

	r.ines2 = flags7&(2<<2) == 2<<2
	r.playChoice10 = flags7&(1<<1) != 0
	r.vsUnisystem = flags7&1 != 0

	prSize := int(r.header[4]) * programRomPageSize
	prStart := headerSize
	if r.hasTrainer {
		prStart += trainerSize
	}
	prEnd := prStart + prSize

	crSize := int(r.header[5]) * charRomPageSize
	crStart := prEnd
	crEnd := crStart + crSize

	if r.hasTrainer {
		trainerStart := headerSize
		trainerEnd := trainerStart + trainerSize

		r.trainer = r.raw[trainerStart:trainerEnd]
	} else {
		r.trainer = []byte{}
	}

	r.programRom = r.raw[prStart:prEnd]
	r.charRom = r.raw[crStart:crEnd]

	return r, nil
}
