package steward

import (
	. "learn/go/concurrency/pattern/or-channel"
	"log"
	"time"
)

// startGoroutineFn is a restart function for goroutines taht can be monitored
// and restarted.
type startGoroutineFn func(
	done <-chan any,
	pulseInterval time.Duration,
) (heartbeat <-chan any)

// newSteward implement the Healing pattern. It monitor the ward which can be
// started/restarted by startGoroutine, timeout is the duration before
// cancelling the ward, if it is inactive (not sending out pulses).
//
// Note that itself return a startGoroutinFn, too. Means that itself is also
// monitorable.
//
// Note a "decorator" pattern here:
//
// * newSteward is a func which receive a func and return a func of the same type.
// * the other parameter 'timeout', is used to configure the returned func.
// * the real word is done by users calling the configured func.
func newSteward(
	timeout time.Duration,
	startGoroutine startGoroutineFn,
) startGoroutineFn {
	return func(
		done <-chan any,
		pulseInterval time.Duration,
	) <-chan any {
		heartbeat := make(chan any)
		go func() {
			defer close(heartbeat)

			var wardDone chan any
			var wardHeartbeat <-chan any

			// startWard wrap the consistent way to start/restart the ward:
			startWard := func() {
				wardDone = make(chan any)
				// We use or-channel pattern here to ensure halting the steward
				// also halt the ward.
				wardHeartbeat = startGoroutine(Or(wardDone, done), timeout/2)
			}
			startWard()

			pulse := time.Tick(pulseInterval)

		monitorLoop:
			for {
				timeoutSignal := time.After(timeout)

				// We need this inner for-loop to ensure steward can send its
				// own pulses.
				for {
					select {
					case <-pulse:
						select {
						case heartbeat <- struct{}{}:
						default:
						}
					case <-wardHeartbeat:
						continue monitorLoop
					case <-timeoutSignal:
						log.Println("steward: ward unhealthy; restarting")
						close(wardDone)
						startWard()
						continue monitorLoop
					case <-done:
						return
					}
				}
			}
		}()

		return heartbeat
	}
}
