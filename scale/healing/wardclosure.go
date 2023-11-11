package steward

import (
	. "learn/go/concurrency/pattern/bridge"
	"log"
	"time"
)

// In the example in steward_test, our ward is a little simplistic: other than
// what’s necessary for cancellation and heartbeats, it takes in no parameters
// and returns no arguments. How might we create a ward that has a shape that
// can be used with our steward? We could:
//
// * Rewrite or generate the steward to fit our wards each time, but this is
//   both cumbersome and unnecessary.
// * Here we use closure.
//
// Let’s take a look at a ward that will generate an integer stream based on a
// discrete list of values.

// doWorkFn return a ward closure that can be monitored by our steward, and a
// channel that the returned ward will be communicating back on.
//
// The ward itself will generate an integer stream based on a discrete list of
// values. We simulate a "unhealthy" behavior in it: when it see negative number,
// it return early.
func doWorkFn(done <-chan any, intList ...int) (startGoroutineFn, <-chan any) {
	// Since we’ll potentially be starting multiple copies of our ward, we make
	// use of bridge channels to help present a single uninterrupted channel to
	// the consumer of doWork.
	intChanStream := make(chan (<-chan any))
	intStream := Bridge(done, intChanStream)
	doWork := func(
		done <-chan any,
		pulseInterval time.Duration,
	) <-chan any {
		intStream := make(chan any)
		heartbeat := make(chan any)
		go func() {
			defer close(intStream)
			select {
			// Our ward closure doWork use the close over intChanStream and
			// intList, but use its own instance's intStream.
			//
			// Notice how intVal flow and bridged together, suppose we've
			// started two instance of doWork 1 and 2:
			//
			//	1: intVal -> instace's intStream -> closed intChanStream \
			//																bridge
			//	2: intVal -> instace's intStream -> closed intChanStream /
			case intChanStream <- intStream:
			case <-done:
				return
			}

			pulse := time.Tick(pulseInterval)

			for {
			valueLoop:
				for _, intVal := range intList {
					if intVal < 0 {
						// The "unhealthy" behavior.
						log.Printf("negative value: %v\n", intVal)
						return
					}

					for {
						select {
						case <-pulse:
							select {
							case heartbeat <- struct{}{}:
							default:
							}
						case intStream <- intVal:
							continue valueLoop
						case <-done:
							return
						}
					}
				}
			}
		}()
		return heartbeat
	}
	return doWork, intStream
}
