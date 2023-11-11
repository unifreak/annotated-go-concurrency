// Cancellation
//
// Cancellation is important because of the net‐work effect: if you’ve begun a
// goroutine, it’s most likely cooperating with several other goroutines in
// some sort of organized fashion. We could even represent this
// interconnectedness as a graph: whether or not a child goroutine should
// continue executing might be predicated on knowledge of the state of many
// other goroutines. goroutine (often the main goroutine) with this full
// contextual knowledge should be able to tell its child goroutines to
// terminate.
package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// TestReadNilLeak show how reading on nil caused goroutine leak.
func TestReadNilLeak(_ *testing.T) {
	doWork := func(strings <-chan string) <-chan interface{} {
		completed := make(chan interface{})
		go func() {
			defer fmt.Println("doWork exited.") // Won't get printed.
			defer close(completed)
			// Reading on nil will block forever, goroutine leak!! the goroutine
			// containing doWork will remain in memeory for the lifetime of
			// this process (in this case is very short).
			for s := range strings {
				fmt.Println(s)
			}
		}()
		return completed
	}

	doWork(nil)
	fmt.Println("Done (but leaked!)")
}

// TestFixReadNilLeakWithDone shows how to cancel goroutine blocked on reading.
func TestFixReadNilLeakWithDone(_ *testing.T) {
	doWork := func(
		done <-chan interface{},
		strings <-chan string, // NOTE: comma required!!
	) <-chan interface{} {
		completed := make(chan interface{})
		go func() {
			defer fmt.Println("doWork exited.")
			defer close(completed)
			for {
				select {
				case <-done:
					return
				case s := <-strings:
					fmt.Println(s)
				}
			}
		}()
		return completed
	}

	done := make(chan interface{})
	// We are still passing in nil, but this time because of doWork's goroutine
	// will check "done", it won't block.
	completed := doWork(done, nil)

	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println("Canceling doWork gorouine...")
		close(done)
	}()

	// Join point, otherwise doWork's goroutine might doesn't have the chance to
	// run.
	<-completed
	fmt.Println("Done.")
}

// TestWritingLeak shows how writing to an channel without reading end causes
// goroutine leak.
func TestWritingLeak(_ *testing.T) {
	newRandStream := func() <-chan int {
		randStream := make(chan int)
		go func() {
			defer fmt.Println("newRandStream closure exited.") // Won't get printed
			defer close(randStream)
			for {
				// After the consumer read 3 ints, because there is no more
				// corresponding read side, here will block forever. goroutine leak!
				randStream <- rand.Int()
			}
		}()
		return randStream
	}

	randStream := newRandStream()
	fmt.Println("3 random ints:")
	for i := 0; i < 3; i++ {
		fmt.Printf("%d: %d\n", i, <-randStream)
	}

	// Simulate ongoing work
	time.Sleep(1 * time.Second)
}

// TestFixWritingLeak shows how to fix writing leak with similar technique: the
// done channel.
func TestFixWritingLeak(_ *testing.T) {
	newRandStream := func(done <-chan interface{}) <-chan int {
		randStream := make(chan int)
		go func() {
			defer fmt.Println("newRandStream closure exited.")
			defer close(randStream)
			for {
				select {
				case <-done:
					return
				case randStream <- rand.Int():
				}
			}
		}()
		return randStream
	}

	done := make(chan interface{})
	randStream := newRandStream(done)
	fmt.Println("3 random ints:")
	for i := 0; i < 3; i++ {
		fmt.Printf("%d: %d\n", i, <-randStream)
	}
	close(done)

	// Simulate ongoing work
	time.Sleep(1 * time.Second)
}

