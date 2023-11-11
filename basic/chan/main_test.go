package main

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"testing"
)

// Note how the resultStream's lifecycle is encapsulated within the chanOwner
// func.
func TestOwnerConsumerResponsibility(t *testing.T) {
											// Owner do:
	chanOwner := func() <-chan int { 	    // * expose with read only chan
		resultStream := make(chan int, 5) 	// * create/encapsulate inside func
		go func() {
			defer close(resultStream) 	    // * close
			for i := 0; i <= 5; i++ {
				resultStream <- i 		  	// * write
			}
		}()
		return resultStream
	}

	resultStream := chanOwner()        			// Consumer do:
	for result := range resultStream { 			// * handle closing/blocking
		fmt.Printf("Received: %d\n", result)	// * read
	}
	fmt.Println("Done receiving")
}

// Perform same behavior as sync.Cond, but channels are composable, this is
// preferred way to unblock multiple goroutines at the same time.
func TestBroadcastByClosing(_ *testing.T) {
	begin := make(chan interface{})
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-begin
			fmt.Printf("%v has begain\n", i)
		}(i)
	}

	fmt.Println("Unblocking goroutines...")
	close(begin) // broadcast
	wg.Wait()
}

func TestBufferWhenKnowSize(_ *testing.T) {
	var stdoutBuff bytes.Buffer
	defer stdoutBuff.WriteTo(os.Stdout)

	intStream := make(chan int, 4)
	go func() {
		defer close(intStream)
		defer fmt.Fprintln(&stdoutBuff, "Producer Done.")
		for i := 0; i < 5; i++ {
			fmt.Fprintf(&stdoutBuff, "Sending: %d\n", i)
			intStream <- i
		}
	}()

	for i := range intStream {
		fmt.Fprintf(&stdoutBuff, "Received %v.\n", i)
	}
}
