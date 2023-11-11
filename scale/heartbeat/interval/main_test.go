package main

import (
	"fmt"
	"testing"
	"time"
)

// doWork implement interval-based heartbeat pattern. It will wait for incoming
// unit of works, process it and send out the result. In the mean time, it also
// sends a pulse every pulseInterval.
func doWork(
	done <-chan any,
	pulseInterval time.Duration,
) (<-chan any, <-chan time.Time) {
	heartbeat := make(chan any)
	results := make(chan time.Time)

	go func() {
		defer close(heartbeat)
		defer close(results)

		pulse := time.Tick(pulseInterval)
		workGen := time.Tick(2*pulseInterval) // simulating incoming works.

		sendPulse := func() {
			select {
			case heartbeat <- struct{}{}:
			default: // guard against blocking when no corresponding receiver.
					 // see comments below.
			}
		}

		sendResult := func(r time.Time) {
			for {
				select {
				case <-done:
					return
				// Just like with done channels, *anytime* we send or receive,
				// we include a case for pulse.
				case <-pulse:
					sendPulse()
				case results <- r:
					fmt.Println("sended r into results...")
					return
				default:
				}
			}

			// @?? Why need the additional `for select {}` loop in sendResult
			//     ()? If I do this (without for loop), it seems also works,
			//     what's the difference?
			//
			// select {
			// case <-done:
			// 	return
			// case <-pulse:
			// 	sendPulse()
			// case results <- r:
			// }

			// @?? Why adding a default case like below will cause r NEVER send
			//     to results?
			//
			// Possibly because of the results reading end is also inside a
			// for-select loop, chance is too small that both the case for
			// reading end and the case for sending end here BOTH got selected?
			//
			// Seems to be so, becuase if we only range over results in test,
			// r will be sent.
			//
			// select {
			// case <-done:
			// 	return
			// case <-pulse:
			// 	sendPulse()
			// case results <- r:
			// 	fmt.Println("sended r into results...")
			// default:	// added default:
			// }
		}

		for {
			select {
			case <-done:
				return
			case <-pulse:
				// heartbeat <- struct{}{}	// possible block forever !!
				//
				// Here is why we cannot simply use the above code, but instead
				// wrap it into a select that have a default case(in sendPulse
				// function): if we do above, when a pulse beats, trying to
				// send to heartbeat channel will block forever if there is no
				// corresponding receiver for heartbeat channel.
				sendPulse()
			case r := <-workGen:
				// results <- r				// possible block forever !!
				//
				// For the same reason, we cannot do above.
				sendResult(r)
			}
		}

	}()

	return heartbeat, results
}

// TestDoWorkHeartbeat shows the usage of doWork's heartbeat.
func TestDoWorkHeartbeat(_ *testing.T) {
	done := make(chan any)
	time.AfterFunc(10*time.Second, func() { close(done) })

	const timeout = 2 * time.Second
	heartbeat, results := doWork(done, timeout/2)

	for {
		select {
		case _, ok := <-heartbeat:
			if ok == false {
				return
			}
			fmt.Println("pulse")
		case r, ok := <-results:
			if ok == false {
				return
			}
			fmt.Printf("results %v\n", r.Second())
		// Here is how timeout interact with heartbeat. If we didn't receive
		// pulse for long enough, we know the goroutine is not healthy, and
		// timeout after a while.
		case <-time.After(timeout):
			fmt.Println("timeout.")
			return
		}
	}
}

// doWork2 also implemente time-interval heartbeat, used to show how the
// heartbeat can help with testing.
func doWork2(
	done <-chan any,
	pulseInterval time.Duration,
	works ...int,
) (<-chan any, <-chan int) {
	heartbeat := make(chan any, 1) 	// @?? buffered
	results := make(chan int)

	go func() {
		defer close(heartbeat)
		defer close(results)

		time.Sleep(2*time.Second)	// simulate delay.

		pulse := time.Tick(pulseInterval)

		loop:
		for _, n := range works {
			for {
				select {
				case <-done:
					return
				case <-pulse:
					select {
					case heartbeat <-struct{}{}:
					default:
					}
				case results <-n:
					continue loop
				}
			}
		}
	}()

	return heartbeat, results
}

// TestDoWork2HeartbeatInTests show how time-interval heartbeat help with
// testing.
//
// The downside is our test logic is a bit muddled. See workbegin/main_test.go.
func TestDoWork2HeartbeatInTests(t *testing.T) {
	done := make(chan any)
	defer close(done)

	ints := []int{0, 1, 2, 3, 5}
	const timeout = 2*time.Second
	heartbeat, results := doWork2(done, timeout/2, ints...)

	<-heartbeat

	i := 0
	for {
		select {
		case r, ok := <-results:
			if !ok {
				return
			} else if expected := ints[i]; r != expected {
				t.Errorf("index %v: expected %v, but received %v,", i, expected, r)
			}
			i++
		case <-heartbeat:
		case <-time.After(timeout):
			t.Fatal("test timed out.")
		}
	}
}