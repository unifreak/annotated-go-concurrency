package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestJoinPointWithWaitGroup(t *testing.T) {
	var wg sync.WaitGroup
	sayHello := func() {
		defer wg.Done()
		fmt.Println("Hello")
	}
	wg.Add(1)
	go sayHello()
	wg.Wait() // join point
}

// TestClosureOpOnCopyTrap shows the trap of how a closure runned in goroutin
// operate on copy of captured variable.
func TestClosureOpOnCopyTrap(t *testing.T) {
	var wg sync.WaitGroup
	salutation := "hello"
	wg.Add(1)
	go func() {
		defer wg.Done()
		salutation = "welcome"
	}()
	wg.Wait()
	fmt.Println(salutation) // "welcome". modified salutation.
}

// TestIterCaptureTrap shows the trap of capturing iteration variables.
func TestIterCaptureTrap(t *testing.T) {
	var wg sync.WaitGroup
	// The Go runtime is observant enough to know that a reference to the
	// salutation variable is still being held, and therefore will transfer the
	// memory to the heap so that the goroutines can continue to access it.
	for _, salutation := range []string{"hello", "greetings", "good day"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println(salutation) // "good day" * 3 times !!!
		}()
	}
	wg.Wait()
}

// TestIterCaptureFix shows how to fix the iteration variable capture trap.
func TestIterCaptureFix(t *testing.T) {
	var wg sync.WaitGroup
	for _, salutation := range []string{"hello", "greetings", "good day"} {
		wg.Add(1)
		go func(salutation string) {
			defer wg.Done()
			fmt.Println(salutation)
		}(salutation)
	}
	wg.Wait()
}
