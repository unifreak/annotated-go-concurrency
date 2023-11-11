package main

import (
	"fmt"
	"math"
	"os"
	"sync"
	"testing"
	"text/tabwriter"
	"time"
)

// Mutex: mutual exclusion.
//
// Mutex provides a concurrent-safe way to express exclusive access to these shared
// resources inside critical section.
//
// Whereas channels share memory by communicating, a Mutex shares memory by
// creating a convention developers must follow to synchronize access to the
// memory. You are responsible for coordinating access to this memory by
// guarding access to it with a mutex.
//
// This code implement a safe counter with mutex.
func TestMutex(_ *testing.T) {
	var count int
	var lock sync.Mutex

	increment := func() {
		lock.Lock()
		defer lock.Unlock()
		count++
		fmt.Printf("Incrementing: %d\n", count)
	}

	decrement := func() {
		lock.Lock()
		defer lock.Unlock()
		count--
		fmt.Printf("Decrementing: %d\n", count)
	}

	var arithmetic sync.WaitGroup
	for i := 0; i < 5; i++ {
		arithmetic.Add(1)
		go func() {
			defer arithmetic.Done()
			increment()
		}()
	}

	for i := 0; i < 5; i++ {
		arithmetic.Add(1)
		go func() {
			defer arithmetic.Done()
			decrement()
		}()
	}

	arithmetic.Wait()
	fmt.Println("Arithmetic complete.")
}

// RWMutex: an arbitrary number of readers can hold a reader lock so long as
// nothing else is holding a writer lock.
//
// This code compare the time cost of mutex and rwmutex in a dummy
// producer/observer pattern.
func TestMutexVsRWMutex(_ *testing.T) {
	// An inactive (imitate by sleeping) producer
	producer := func(wg *sync.WaitGroup, l sync.Locker) {
		defer wg.Done()
		for i := 5; i > 0; i-- {
			l.Lock()
			l.Unlock()
			time.Sleep(1)
		}
	}

	// An active observer
	observer := func(wg *sync.WaitGroup, l sync.Locker) {
		defer wg.Done()
		l.Lock()
		defer l.Unlock()
	}

	// test the time cost of the combination of 1 producer + N(count) observer.
	//
	// sync.Locker is an interface having Lock and Unlock methods. Both Mutex
	// and RWMutex satisfy it.
	test := func(count int, mutex, rwmutex sync.Locker) time.Duration {
		var wg sync.WaitGroup
		wg.Add(count + 1)
		beginTestTime := time.Now()
		go producer(&wg, mutex)
		for i := count; i > 0; i-- {
			go observer(&wg, rwmutex)
		}

		wg.Wait()
		return time.Since(beginTestTime)
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 1, 2, ' ', 0)
	defer tw.Flush()

	var m sync.RWMutex
	fmt.Fprintf(tw, "Readers\tRWMutex\tMutex\n")
	for i := 0; i < 20; i++ {
		count := int(math.Pow(2, float64(i)))
		fmt.Fprintf(
			tw,
			"%d\t%v\t%v\n",
			count,

			// m.RLocker will return a "read-only" locker, which means after the
			// following code, calling r.Lock is the same as calling rw.RLock
			// (and the same goes for Unlock):
			//
			// 		var rw RWMutex;
			// 		r := rw.RLocker();
			//
			// making the l.Lock in observer essentially request a read lock.
			test(count, &m, m.RLocker()), // observer with RWMutex
			// without m.RLocker(), l.Lock in observer request a write lock.
			test(count, &m, &m),          // observer with Mutex
		)
	}
}

