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
	"github.com/evandigby/nesgo/nes"
	"github.com/evandigby/nesgo/ppu"
	"github.com/evandigby/nesgo/rom"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() - 1)
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

	exit := make(chan bool)

	renderer := ppu.NewWebSocketRenderer("/play")
	p := ppu.NewPPU(ppuchan, ines, renderer)

	n := nes.NewNES(p.MemoryMap)

	n.LoadRom(ines)

	nesCPU := cpu.NewCPU(n, exit, cpuLog, nesLog)

	clock := clock.NewClock(21477272, nesCPU.Sync, ppuchan)
	if cpuLog == nil {
		//clock.Pause()
	}
	clock.Run()
	nesCPU.Run()
	p.Run()

	go func() {
		<-exit
		clock.Stop()
	}()

	/*
		rand.Seed(time.Now().Unix())
		go func() {
			colors := [][]byte{
				[]byte{0xFF, 0x00, 0x00, 0xFF},
				[]byte{0x00, 0xFF, 0x00, 0xFF},
				[]byte{0x00, 0x00, 0xFF, 0xFF},
			}

			for { //_ = range time.NewTicker(time.Second).C {
				img := image.NewNRGBA(image.Rect(0, 0, 256, 240))
				for i := 0; i < len(img.Pix); i += 4 {
					color := 0
					if rand.Int31n(100) < 10 {
						color = rand.Intn(len(colors))
					}
					img.Pix[i] = colors[color][0]
					img.Pix[i+1] = colors[color][1]
					img.Pix[i+2] = colors[color][2]
					img.Pix[i+3] = colors[color][3]
				}
				renderer.Render(img)
			}
		}()
	*/

	debugger := debug.NewDebugger(n, nesCPU, p, clock, "./debug/ui/")
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
