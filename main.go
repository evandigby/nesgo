package main

import (
	"fmt"
	"os"

	"github.com/evandigby/nesgo/cpu"
	"github.com/evandigby/nesgo/rom"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Not enough args\n")
		return
	}

	file, err := os.Open(os.Args[1])

	if err != nil {
		fmt.Printf("Unable to open file %v\n", os.Args[1])
		return
	}
	defer file.Close()

	i, err := rom.NewINES(file)

	if err != nil {
		fmt.Printf("Error opening iNES File: %v\n", err)
		return
	}

	d := cpu.Decompile(i.ProgramRom())

	for i, o := range d {
		fmt.Printf("%X: %v\n", 0x8000+i, o.Disassemble())
	}
}
