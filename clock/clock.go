package clock

import (
	"fmt"
	"sync"
	"time"
)

type Clock struct {
	frequency uint64
	cpu       chan int
	ppu       chan int
	tick      uint64
	running   bool
	isPaused  bool
	pause     chan bool
	isDone    bool
	done      chan bool
	dl        sync.Mutex
}

func NewClock(frequency uint64, cpu chan int, ppu chan int) *Clock {
	return &Clock{frequency, cpu, ppu, 0, false, false, make(chan bool), true, make(chan bool), sync.Mutex{}}
}

func waitFor(ch chan bool) {
	for {
		v := <-ch
		if v {
			return
		}
	}
}

func (c *Clock) execute() {
	c.isDone = false
	startTime := time.Now()
	interval := time.Second / time.Duration(c.frequency)
	fmt.Printf("Started Clock at %v with interval %v (%vMHz)\n", startTime, interval, float64(c.frequency)/1000000.0)

	for {
		if c.isDone {
			break
		} else if c.isPaused {
			<-c.pause
		}

		c.cpu <- 0
		cycles := <-c.cpu
		totalCycles := cycles * 3
		for i := 0; i < totalCycles; i++ {
			c.ppu <- 0
			<-c.ppu
		}
		c.tick += uint64(totalCycles)
	}
	endTime := time.Now()

	totalTime := endTime.Sub(startTime)
	seconds := totalTime / time.Second
	if seconds == 0 {
		seconds = 1
	}
	mhz := float64(c.tick/uint64(seconds)) / 1000000.0
	fmt.Printf("Stopped Clock at %v after %v ticks for %vMHz\n", endTime, c.tick, mhz)
	c.running = false
	c.done <- true
}

func (c *Clock) Run() {
	if c.running {
		return
	}

	c.running = true
	go c.execute()
}

func (c *Clock) Pause() {
	c.isPaused = true
}

func (c *Clock) Resume() {
	c.isPaused = false
	c.pause <- false
}

func (c *Clock) Step() {
	if c.isPaused {
		c.pause <- true
	}
}

func (c *Clock) Stop() {
	if c.running {
		c.isDone = true
		if c.isPaused {
			c.pause <- true
		}
		<-c.done
	}
}
