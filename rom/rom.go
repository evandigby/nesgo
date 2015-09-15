package rom

type ROM interface {
	Trainer() []*byte
	Pages() int
	ProgramRom() []*byte
	CharRom() []*byte
	PlayChoiceInstRom() []*byte
	PlayChoicePRom() []*byte
}
