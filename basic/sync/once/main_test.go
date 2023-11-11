package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestDoOnce(_ *testing.T) {
	var count int

	increment := func() {
		count++
	}

	var once sync.Once

	var increments sync.WaitGroup
	increments.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer increments.Done()
			once.Do(increment)
		}()
	}

	increments.Wait()
	fmt.Printf("Count is %d\n", count) // "1"
}

func TestDoTwoFuncOnce(_ *testing.T) {
	var count int
	increment := func() { count++ }
	decrement := func() { count-- }

	var once sync.Once
	once.Do(increment)
	once.Do(decrement)

	fmt.Printf("Count is %d\n", count) // "1" !
}

// The only thing sync.Once guarantees is that your functions are only called once.
// Sometimes this is done by deadlocking your program and exposing the flaw in your
// logic -- in this case a circular reference.
func TestDoOnceCircularDeadlock(_ *testing.T) {
	var onceA, onceB sync.Once
	var initB func()
	initA := func() { onceB.Do(initB) }
	// Do won't proceed until the call to Do on next line exits. A classic
	// example of deadlock:
	//
	// 		onceA					onceB
	// 		-----------------------------------
	// 		initA		---->		initB
	// 					<-x--
	initB = func() { onceA.Do(initA) }

	onceA.Do(initA) // deadlock!
}

