package main

import (
	"fmt"
	"testing"
	"time"
)

// doWork implements the "Heartbeat when work start" pattern. It send a pulse
// before every processing of work unit.
func doWork(done <-chan any, works ...int) (<-chan any, <-chan int) {
	// Because our heartbeat channel was created with a buffer of one, if
	// someone is listening, but not in time for the first pulse, they’ll still
	// be notified of a pulse, hence break our promise to send pulse
	// before *every* unit of works.
	//
	// Suppose we are using unbuffered channel, here is the situation might
	// happens (if the receiver side of heartbeat doesn't setup early enough to
	// receive the first pulse):
	//
	// 		doWork							receiver
	//		-----------------------------   ------------------------------------
	// 		unbuffered<-pulse				(not setup yet)
	// 			(will pass through default)
	// 										<-heartbeat	(missed the first pulse)
	//		results<-i						<-results	("i")
	//		unbuffered<-pulse							("pulse")
	//		results<-i									("i")
	//		...								...
	//
	heartbeat := make(chan any, 1)
	results := make(chan int)

	go func() {
		defer close(heartbeat)
		defer close(results)

		// Simulate delay caused by CPU load, disk contention, network latency,
		// goblins... before the goroutine can begin working.
		time.Sleep(2 * time.Second)

		for _, n := range works {
			// Here we set up a separate select block for the heartbeat. We
			// don’t want to include this in the same select block as the send
			// on results because if the receiver isn’t ready for the result,
			// they’ll receive a pulse instead, and the current value of the
			// result *will be lost*.
			//
			// We also don’t include a case statement for the done channel since
			// we have a default case that will just fall through.
			select {
			case heartbeat <- struct{}{}:
			// Once again we guard against the fact that no one may be listening
			// to our heartbeats.
			default:
			}

			select {
			case <-done:
				return
			case results <- n:
			}
		}
	}()
	return heartbeat, results
}

// TestDoWorkHeartbeat shows the usage of doWork's heartbeat.
func TestDoWorkHeartbeat(_ *testing.T) {
	done := make(chan any)
	defer close(done)

	works := []int{1, 3, 5, 7}
	heartbeat, results := doWork(done, works...)
	for {
		select {
		case _, ok := <-heartbeat:
			if !ok {
				return
			}
			fmt.Println("pulse")
		case r, ok := <-results:
			if !ok {
				return
			}
			fmt.Printf("results %v\n", r)
		}
	}
}

// TestBadTimeoutTests show a BAD test practice of how timing out makes our
// tests nondeterministic.
func TestBadTimeoutTests(t *testing.T) {
	done := make(chan any)
	defer close(done)

	ints := []int{0, 1, 2, 3, 5}
	_, results := doWork(done, ints...)
	for i, expected := range ints {
		select {
		case r := <-results:
			if r != expected {
				t.Errorf(
					"index %v: expected %v, but received %v,",
					i,
					expected,
					r,
				)
			}
		// PROBLEMATIC!! we try to prevent a broken goroutine from deadlocking
		// our test by timing out, but this make our tests not deterministic,
		// it might fail or success due how long will the working goroutine
		// delay before it give results.
		case <-time.After(1*time.Second):
			t.Fatal("test timed out")
		}
	}
}

// TestFixBadTimeoutTest shows how heartbeat help fixing our nondeterministc
// tests.
//
// But it has one risk: the first iteration in our test take inordinate long
// time, due to waiting for the first heartbeat. If this is important for us,
// we can use interval-based heartbeat instead. See interval/main_test.go.
//
// If we're reasonably sure the goroutine’s loop won’t stop executing once it’s
// started, prefer only blocking on the first heartbeat and then falling into a
// simple range statement (this, instead of time-interval based). We can write
// separate tests that specifically test for failing to close channels, loop
// iterations taking too long, and any other timing-related issues.
func TestFixBadTimeoutTest(t *testing.T) {
	done := make(chan any)
	defer close(done)

	ints := []int{0, 1, 2, 3, 5}
	heartbeat, results := doWork(done, ints...)

	<-heartbeat	// work begin.

	i := 0
	for r := range results {
		if expected := ints[i]; r != expected {
			t.Errorf("index %v: expected %v, but received %v,", i, expected, r)
		}
		i++
	}
}