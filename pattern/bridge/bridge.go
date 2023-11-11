package bridge

import (
	. "learn/go/concurrency/pattern/ordone"
)

// bridge implement the Bridge Channel pattern.
func Bridge(
	done <-chan interface{},
	chanStream <-chan <-chan interface{},
) <-chan interface{} {
	valStream := make(chan interface{})
	go func() {
		defer close(valStream)
		for stream := range chanStream {
			select {
			case <-done:
				return
			default:
			}

			for v := range OrDone(done, stream) {
				select {
				case <-done:
				case valStream <- v:
				}
			}
		}
	}()
	return valStream
}
