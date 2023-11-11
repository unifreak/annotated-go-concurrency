package main

import (
	"fmt"
	"sync"
	"testing"
)

// Pool’s primary interface is its Get method. When called, Get will first check
// whether there are any available instances within the pool to return to the
// caller, and if not, call its New member variable to create a new one. When
// finished, *callers* call Put to place the instance they were working with
// back in the pool for use by other processes.
func TestPool(_ *testing.T) {
	myPool := &sync.Pool{
		New: func() interface{} {
			fmt.Println("Creating new instance.")
			return struct{}{} // dummy instance
		},
	}

	myPool.Get()             // will call New. since no instance in pool
	instance := myPool.Get() // will call New. since no instance in pool
	myPool.Put(instance)     // now there is one instance in pool
	myPool.Get()             // won't call New
}

// This code shows using Pool to restrict memory usage.
//
// Another useage of Pool is warming up memory, See warm_test.go
func TestSeedPoolToSaveMemory(_ *testing.T) {
	var numCalcsCreated int
	calcPool := &sync.Pool{
		New: func() interface{} {
			numCalcsCreated++
			mem := make([]byte, 1024) // 1kb
			return &mem
		},
	}

	// Seed the pool with 4kb
	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())

	const numWorkers = 1024 * 1024
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for i := numWorkers; i > 0; i-- {
		go func() {
			defer wg.Done()

			mem := calcPool.Get().(*[]byte) // asserting a pointer to a slice of bytes
			defer calcPool.Put(mem)

			// ...some quick processing with this memory...
		}()
	}

	wg.Wait()

	// NOTE the num of created calc is non-deterministic. Even though, without
	// Pool, in the worst case we could have been attempting to allocate a
	// gigabyte of memory.
	fmt.Printf("%d calculator were created.\n", numCalcsCreated)
}
