package main

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

// doWork is our worker that may have different delay doing work due to
// different loads, network latency...
func doWork(done <-chan any, id int, wg *sync.WaitGroup, result chan<- int) {
	started := time.Now()
	defer wg.Done()

	// Simulate random work delay.
	simulatedLoadTime := time.Duration(1+rand.Intn(5)) * time.Second
	select {
	case <-done:
	case <-time.After(simulatedLoadTime):
	}

	select {
	case <-done:
	case result <- id:
	}

	took := time.Since(started)
	if took < simulatedLoadTime { // might cancelled early by done channel.
		took = simulatedLoadTime
	}
	fmt.Printf("%v took %v\n", id, took)
}

// TestReplicatedDoWork shows how replicated requests pattern works.
func TestReplicatedDoWork(_ *testing.T) {
	done := make(chan any)
	result := make(chan int)

	var wg sync.WaitGroup
	wg.Add(10)

	// We replicate work to be handled by 10 workers.
	for i := 0; i < 10; i++ {
		go doWork(done, i, &wg, result)
	}

	firstReturned := <-result
	close(done)
	wg.Wait()

	fmt.Printf("Received an answer from #%v\n", firstReturned)
}
