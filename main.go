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
	runtime.GOMAXPROCS(4)
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

	var cpuLog, nesLog *os.File
	if len(os.Args) > 2 {
		cpuLog, err = os.Create(os.Args[2])
		if err != nil {
			fmt.Printf("Unable to create log %v\n", err)
			return
		}
		defer cpuLog.Close()
	}

	nesLog, err = os.Open("nestest.log")

	if err != nil {
		fmt.Printf("Unable to open nestest.log %v\n", err)
		nesLog = nil
	} else {
		defer nesLog.Close()
	}

	ppu := make(chan int)

	go func() {
		for {
			ppu <- 0
		}
	}()

	exit := make(chan bool)

	nesCPU := cpu.NewCPU(ines, exit, cpuLog, nesLog)
	clock := clock.NewClock(21477272, nesCPU.Sync, ppu)
	if cpuLog == nil {
		clock.Pause()
	}
	clock.Run()
	nesCPU.Run()

	go func() {
		<-exit
		clock.Pause()
	}()

	debugger := debug.NewDebugger(nesCPU.State, clock, "./debug/ui/")
	debugger.Start()

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			switch text {
			case "quit\n":
				clock.Stop()
				return
			case "step\n", "\n":
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
