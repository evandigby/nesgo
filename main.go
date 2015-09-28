package main

import (
	"bufio"
	"fmt"
	"image"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/evandigby/nesgo/clock"
	"github.com/evandigby/nesgo/cpu"
	"github.com/evandigby/nesgo/debug"
	"github.com/evandigby/nesgo/ppu"
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

	ppuchan := make(chan int)

	go func() {
		for {
			ppuchan <- 0
		}
	}()

	exit := make(chan bool)

	nesCPU := cpu.NewCPU(ines, exit, cpuLog, nesLog)
	clock := clock.NewClock(21477272, nesCPU.Sync, ppuchan)
	if cpuLog == nil {
		clock.Pause()
	}
	clock.Run()
	nesCPU.Run()

	go func() {
		<-exit
		clock.Stop()
	}()

	renderer := ppu.NewWebSocketRenderer("/play")
	rand.Seed(time.Now().Unix())
	go func() {
		img := image.NewNRGBA(image.Rect(0, 0, 256, 240))
		red := true

		for { //_ = range time.NewTicker(time.Second).C {
			for i := range img.Pix {
				if i%3 == 0 {
					img.Pix[i] = 0xFF
					continue
				}
				img.Pix[i] = byte(rand.Int())
			}
			/*
				for i := 0; i < len(img.Pix); i += 4 {
					if red {
						img.Pix[i] = 0xFF
						img.Pix[i+1] = 0x00
					} else {
						img.Pix[i] = 0x00
						img.Pix[i+1] = 0xFF
					}
					img.Pix[i+3] = 0xFF
				}
			*/
			red = !red
			renderer.Render(img)
		}
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
