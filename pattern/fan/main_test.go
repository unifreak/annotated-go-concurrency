// Try to find prime numbers in a series of random numbers, using a slow stage.
package main

import (
	"fmt"
	. "learn/go/concurrency/pattern/pipeline/stage"
	mathrand "math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

func rand() interface{} {
	return mathrand.Intn(50000000)
}

// primeFinder try to find prime numbers. It's the slow stage we deliberately
// set up, to show how fan-out fan-in help with pipeline performance.
func primeFinder(
	done <-chan interface{},
	intStream <-chan int,
) <-chan interface{} {
	isPrime := func(v int) bool { // SLOW!
		for i := 2; i < v; i++ {
			if v%i == 0 {
				return false
			}
		}
		return true
	}

	primeStream := make(chan interface{})
	go func() {
		defer close(primeStream)
		for v := range intStream {
			select {
			case <-done:
				return
			default:
			}

			if isPrime(v) {
				primeStream <- v
			}
		}
	}()
	return primeStream
}

func TestSlowPipeline(_ *testing.T) {
	done := make(chan interface{})
	defer close(done)

	start := time.Now()

	randIntStream := ToInt(done, RepeatFn(done, rand))
	fmt.Println("Primes:")
	for prime := range Take(done, primeFinder(done, randIntStream), 10) {
		fmt.Printf("\t%d\n", prime)
	}
	fmt.Printf("Search took: %v\n", time.Since(start))
}

func TestFanOutFanInPipeline(_ *testing.T) {
	done := make(chan interface{})
	defer close(done)

	start := time.Now()

	randIntStream := ToInt(done, RepeatFn(done, rand))

	// Fan-Out
	//
	// The process of fanning out a stage in a pipeline is extraordinarily easy,
	// all we have to do is start multiple versions of that stage.
	//
	// In production, we would probably do a little empirical testing to
	// determine the optimal number of CPUs, but here we’ll stay simple and
	// assume that a CPU will be kept busy by only one copy of the findPrimes
	// stage.
	numFinders := runtime.NumCPU()
	fmt.Printf("Spinning up %d prime finders.\n", numFinders)
	finders := make([]<-chan interface{}, numFinders)
	for i := 0; i < numFinders; i++ {
		finders[i] = primeFinder(done, randIntStream)
	}

	// Fan-In
	//
	// Now that we have four goroutines, we also have four channels, but our
	// range over primes is only expecting one channel. This brings us to the
	// fan-in portion of the pattern.
	//
	// Preserving Orders
	//
	// A naive implementation of the fan-in, fan-out algorithm (like our example
	// heree) only works if the order in which results arrive is unimportant.
	// We have done nothing to guarantee that the order in which items are read
	// from the randIntStream is preserved as it makes its way through the
	// prime sieve stage.
	fanIn := func(
		done <-chan interface{},
		channels ...<-chan interface{},
	) <-chan interface{} {
		// In a nutshell, fanning in implementation involves creating the
		// multiplexed(joined together) channel that consumers will read from,
		// then spinning up one goroutine for each incoming channel, and one
		// goroutine to close the multiplexed channel when the incoming
		// channels have all been closed. Since we’re going to be creating a
		// goroutine that is waiting on N other goroutines to complete, it
		// makes sense to create a sync.WaitGroup to coordinate things. The
		// multiplex function also notifies the WaitGroup that it’s done.

		// Use wg to wait until all channels have been drained.
		var wg sync.WaitGroup

		multiplexedStream := make(chan interface{})
		multiplex := func(c <-chan interface{}) {
			defer wg.Done()
			for i := range c {
				select {
				case <-done:
					return
				case multiplexedStream <- i:
				}
			}
		}

		// Select from all the channels.
		wg.Add(len(channels))
		for _, c := range channels {
			go multiplex(c)
		}

		// Wait for all the reads to complete.
		go func() {
			wg.Wait()
			close(multiplexedStream)
		}()

		return multiplexedStream
	}

	fmt.Println("Primes:")
	for prime := range Take(done, fanIn(done, finders...), 10) {
		fmt.Printf("\t%d\n", prime)
	}
	fmt.Printf("Search took: %v\n", time.Since(start))
}
