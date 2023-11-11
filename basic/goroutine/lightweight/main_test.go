// This program calculate goroutine size.
package main

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
)

func TestSize(_ *testing.T) {
	memConsumed := func() uint64 {
		runtime.GC()
		var s runtime.MemStats
		runtime.ReadMemStats(&s)
		return s.Sys
	}

	var c <-chan interface{}
	var wg sync.WaitGroup

	// Since gc does nothing to collect coroutines that have been abandoned
	// somehow, we can deliberatelly caused goroutine leak like this for
	// measurement.
	noop := func() { wg.Done(); <-c }

	const numGoroutines = 1e4
	wg.Add(numGoroutines)
	before := memConsumed()
	for i := numGoroutines; i > 0; i-- {
		go noop()
	}

	wg.Wait()
	after := memConsumed()
	fmt.Printf("%.3fkb\n", float64(after-before)/numGoroutines/1000)
}

// run: go test -bench=. -cpu=1
//
// compare this with linux thread switching:
//
// 		taskset -c 0 perf bench sched pipe -T
func BenchmarkContextSwitch(b *testing.B) {
	var wg sync.WaitGroup
	begin := make(chan struct{})
	c := make(chan struct{})

	var token struct{}
	sender := func() {
		defer wg.Done()
		<-begin // wait until told to begin, factor out cost of setups
		for i := 0; i < b.N; i++ {
			c <- token // send
		}
	}

	receiver := func() {
		defer wg.Done()
		<-begin // same as sender
		for i := 0; i < b.N; i++ {
			<-c // receive
		}
	}

	wg.Add(2)
	go sender()
	go receiver()
	close(begin) // tell sender/receiver to begin
	wg.Wait()
}

