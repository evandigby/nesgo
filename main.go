package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"sync"

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

	ppu := make(chan int)

	go func() {
		for {
			ppu <- 0
		}
	}()

	nesCPU := cpu.NewCPU(ines)
	clock := clock.NewClock(21477272, nesCPU.Sync, ppu)
	clock.Pause()
	clock.Run()
	nesCPU.Run()

	debugger := debug.NewDebugger(nesCPU.State, clock, "./debug/ui/")
	debugger.Start()

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("> ")
			text, _ := reader.ReadString('\n')
			switch text {
			case "quit\n":
				clock.Stop()
				return
			case "step\n":
				clock.Step()
			case "pause\n":
				clock.Pause()
			case "resume\n":
				clock.Resume()
			}
		}
	}()

	wg.Wait()
}
