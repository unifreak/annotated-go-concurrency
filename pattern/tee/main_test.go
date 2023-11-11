package main

import (
	"fmt"
	. "learn/go/concurrency/pattern/ordone"
	. "learn/go/concurrency/pattern/pipeline/stage"
	"testing"
)

// tee implements the tee-channel pattern.
func tee(
	done <-chan interface{},
	in <-chan interface{},
) (_, _ <-chan interface{}) {
	out1 := make(chan interface{})
	out2 := make(chan interface{})
	go func() {
		defer close(out1)
		defer close(out2)
		for v := range OrDone(done, in) {
			// We shadow out1 and out2 because we need to set them to nil
			// after sending one value into them. See below.
			var out1, out2 = out1, out2

			// We use one select statement so that writes to out1 and out2
			// don’t block each other.
			//
			// To ensure both are written to, we’ll perform "two" iterations
			// of the select statement: one for each outbound channel.
			for i := 0; i < 2; i++ {
				select {
				case <-done:
					return
				// Once one of them is written, it's set to nil, so the next
				// iteration will have to write to the other. Hence we got
				// both out1 and out2 written.
				case out1 <- v:
					out1 = nil
				case out2 <- v:
					out2 = nil
				}
			}
		}
	}()
	return out1, out2
}

func TestTee(_ *testing.T) {
	done := make(chan interface{})
	defer close(done)

	out1, out2 := tee(done, Take(done, Repeat(done, 1, 2), 4))
	for v1 := range out1 {
		fmt.Printf("out1: %v, out2: %v\n", v1, <-out2)
	}
}
