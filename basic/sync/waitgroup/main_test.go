// WaitGroup is a great way to wait for a set of concurrent operations to complete
// when you:
//
// * either don’t care about the result of the concurrent operation
// * or you have other means of collecting their results.
//
// If neither of those conditions are true, I suggest you use channels and a select
// statement instead.
//
// You can think of a WaitGroup like a concurrent-safe counter: calls to Add
// increment the counter by the integer passed in, and calls to Done decrement the
// counter by one. Calls to Wait block until the counter is zero.
package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestTrackEach(_ *testing.T) {
	var wg sync.WaitGroup

	// Notice that the calls to Add are done outside the goroutines they’re
	// helping to track.
	//
	// Had the calls to Add been placed inside the goroutines’ closures, the
	// call to Wait could have returned without blocking at all because the
	// calls to Add would not have taken place (a race condition!).

	wg.Add(1) // 1 goroutine is beginning
	go func() {
		defer wg.Done() // goroutine exit
		fmt.Println("1st goroutine sleeping...")
		time.Sleep(1)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("2nd goroutin sleeping...")
		time.Sleep(2)
	}()

	wg.Wait() // block until all goroutines exit
	fmt.Println("all goroutines complete.")
}

func TestTrackBatch(_ *testing.T) {
	hello := func(wg *sync.WaitGroup, id int) {
		defer wg.Done()
		fmt.Printf("hello from %v\n", id)
	}

	// It’s customary to couple calls to Add as closely as possible to the
	// goroutines they’re helping to track, but sometimes you’ll find Add
	// called to track a group of goroutines all at once. I usually do this
	// before for loops like this:
	const numGreeters = 5
	var wg sync.WaitGroup
	wg.Add(numGreeters)
	for i := 0; i < numGreeters; i++ {
		go hello(&wg, i+1)
	}

	wg.Wait()
}

