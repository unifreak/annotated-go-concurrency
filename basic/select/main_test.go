package main

import (
	"fmt"
	"testing"
	"time"
)

func TestBlockWhenNoneReady(_ *testing.T) {
	start := time.Now()
	c := make(chan interface{})
	go func() {
		time.Sleep(5 * time.Second)
		close(c)
	}()

	// if none of the channels are ready, the entire select statement blocks.
	fmt.Println("Blocking on read...")
	select {
	case <-c:
		fmt.Printf("Unblocked %v later.\n", time.Since(start))
	}
}

func TestEqualChanceOfSelection(_ *testing.T) {
	c1 := make(chan interface{})
	close(c1)
	c2 := make(chan interface{})
	close(c2)

	var c1Count, c2Count int
	for i := 0; i < 1000; i++ {
		select {
		case <-c1:
			c1Count++
		case <-c2:
			c2Count++
		}
	}

	fmt.Printf("c1Count: %d\nc2Count: %d\n", c1Count, c2Count)
}

// Ways to handle channel may be never ready
//
// * time out
// * break wait with default case
// * mix: continue work but also check readiness

func TestTimeout(_ *testing.T) {
	var c <-chan int
	select {
	case <-c:
	// After() return a channel that send current time after specified interval.
	case <-time.After(1 * time.Second):
		fmt.Println("Timed out.")
	}
}

func TestBreakWithDefault(_ *testing.T) {
	start := time.Now()

	var c1, c2 chan int
	select {
	case <-c1:
	case <-c2:
	default:
		fmt.Printf("In default after %v\n", time.Since(start))
	}
}

// Usually we’ll see a default clause used in conjunction with a for-select
// loop. This allows a goroutine to make progress on work while waiting for
// another goroutine to report a result.
func TestForDefault(_ *testing.T) {
	done := make(chan interface{})
	go func() {
		time.Sleep(5 * time.Second)
		close(done)
	}()

	workCounter := 0
loop:
	for {
		// Check
		select {
		case <-done:
			break loop
		default:
		}

		// Simulate work
		workCounter++
		time.Sleep(1 * time.Second)
	}

	fmt.Printf("Achieved %v cycles of work before signalled to stop.\n", workCounter)
}

