package stage

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestTakeFromRepeat(_ *testing.T) {
	done := make(chan interface{})
	defer close(done)

	// Repeat generator is very efficient here, because the Repeat generator's
	// send blocks on the Take stage's receive, Although Repeat have the
	// capability of generating an infinite stream of ones, it only generate
	// N+1 instances where N is the number we pass into the Take stage.
	for v := range Take(done, Repeat(done, 1), 10) {
		fmt.Printf("%v ", v)
	}
	fmt.Println()
}

func TestTakeFromRepeatFn(_ *testing.T) {
	done := make(chan interface{})
	defer close(done)

	rand := func() interface{} { return rand.Int() }
	for num := range Take(done, RepeatFn(done, rand), 10) {
		fmt.Println(num)
	}
}

func TestToString(_ *testing.T) {
	done := make(chan interface{})
	defer close(done)

	var message string
	for token := range ToString(done, Take(done, Repeat(done, "I", "am."), 5)) {
		message += token
	}

	fmt.Println("message:", message)
}
