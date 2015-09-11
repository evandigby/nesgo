package rom

type ROM interface {
	Trainer() []byte
	ProgramRom() []byte
	CharRom() []byte
	PlayChoiceInstRom() []byte
	PlayChoicePRom() []byte
}
