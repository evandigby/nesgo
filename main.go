package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/evandigby/nesgo/clock"
	"github.com/evandigby/nesgo/cpu"
	"github.com/evandigby/nesgo/debug"
	"github.com/evandigby/nesgo/rom"
)

func main() {
	old := runtime.GOMAXPROCS(4)
	fmt.Printf("Old: %v\n", old)
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

	ines, err := rom.NewINES(file)

	if err != nil {
		fmt.Printf("Error opening iNES File: %v\n", err)
		return
	}

	d := cpu.Decompile(ines.ProgramRom())
	/*
		i := 0
		for {
			if i >= len(d) {
				break
			}

			o := d[i]
			fmt.Printf("%X: %v\n", 0x8000+i, o.Disassemble())

			i += len(o.Opcode())
		}
	*/
	e := cpu.Execution(d)
	s := cpu.NewState()

	ppu := make(chan int)

	go func() {
		for {
			ppu <- 0
		}
	}()

	go func() {
		for {
			for _, e := range e {
				c, _ := e(s)
				s.Sync <- c
			}
		}
	}()

	clock := clock.NewClock(21477272, s.Sync, ppu)
	clock.Run()

	debugger := debug.NewDebugger(s, clock, "./debug/ui/")
	debugger.Start()
	time.Sleep(100 * time.Second)
	clock.Stop()
}
