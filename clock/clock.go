package clock

import (
	"fmt"
	"sync"
	"time"
)

type ClockOperation struct {
	Tick    func(uint64)
	Exiter  func()
	Divisor int64
}

func NewClockOperation(divisor int64, tick func(uint64), exiter func()) *ClockOperation {
	return &ClockOperation{tick, exiter, divisor}
}

type Clock struct {
	frequency  uint64
	operations []*ClockOperation
	tick       uint64
	running    bool
	isPaused   bool
	pause      chan bool
	isDone     bool
	done       chan bool
	dl         sync.Mutex
}

func NewClock(frequency uint64, operations ...*ClockOperation) *Clock {
	return &Clock{frequency, operations, 0, false, false, make(chan bool), true, make(chan bool), sync.Mutex{}}
}

func waitFor(ch chan bool) {
	for {
		v := <-ch
		if v {
			return
		}
	}
}

func (c *Clock) exit() {
	for _, o := range c.operations {
		o.Exiter()
	}
}

func (c *Clock) execute() {
	//divisorStates := make([]int64, len(c.operations))
	c.isDone = false
	startTime := time.Now()
	//	ticker := time.NewTicker(interval)

	//	defer func() {
	//	}()

	o := c.operations[0].Tick
	d := 0
	for { // range ticker.C {
		/*for i, o := range c.operations {
			s := divisorStates[i]
			if s == o.Divisor {
				divisorStates[i] = 0
				o.Tick(c.tick)
			}
			divisorStates[i]++
		}
		*/
		if d == 12 {
			o(c.tick)
			d = 0
		}
		d++

		c.tick++

		if c.isDone {
			break
		} else if c.isPaused {
			waitFor(c.pause)
		}
	}

	interval := time.Second / time.Duration(c.frequency)
	fmt.Printf("Started Clock at %v with interval %v (%vMHz)\n", startTime, interval, float64(c.frequency)/1000000.0)

	endTime := time.Now()
	//ticker.Stop()
	go c.exit()
	totalTime := endTime.Sub(startTime)
	seconds := totalTime / time.Second
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
	c.pause <- false
}

func (c *Clock) Stop() {
	if c.running {
		c.isDone = true
		<-c.done
	}
}
