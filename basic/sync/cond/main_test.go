package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// This code shows unicasting with Signal().
//
// Here we have a queue of fixed length 2, and 10 items we want to push onto the
// queue. We want to enqueue items as soon as there is room, so we want to be
// notified as soon as there’s room in the queue. We can use Cond to enqueue
// items as soon as there is room.
func TestUnicastBySignal(_ *testing.T) {
	c := sync.NewCond(&sync.Mutex{})
	queue := make([]interface{}, 0, 10)

	removeFromQueue := func(delay time.Duration) {
		time.Sleep(delay)
		c.L.Lock()
		queue = queue[1:]
		fmt.Println("Removed from queue")
		c.L.Unlock()
		c.Signal() // notify goroutine blocked on Wait()
	}

	for i := 0; i < 10; i++ {
		c.L.Lock()

		// Note here we check the length of the queue in a loop. This is
		// important because a signal on the condition doesn’t necessarily mean
		// what you’ve been waiting for has occurred, only that *something* has
		// occurred.
		//
		// Suppose this is the third iteration (first entrance in Wait), here is
		// what happend regarding to lock:
		//
		//		Producer			Wait()							Consumer
		//		-------------------------------------------------------------------
		//		Lock() 	    -->		Unlock()
		//				^	  	  	  allow other processing
		//				 \	 	  	  and signal
		//				  \	_ _ 	  	  						-->	Lock()
		//						\  	  								...processing...
		//		(back to Lock)   \ 		  							Unlock()
		//			check room		Lock()						<--	Signal()
		//			to continue
		//			producing
		//
		for len(queue) == 2 { // the condition we need to check for
			c.Wait()		  // wait until at least one item is dequeued
		}
		fmt.Println("Adding to queue")
		queue = append(queue, struct{}{})
		go removeFromQueue(1 * time.Second)
		c.L.Unlock()
	}
	// NOTE will exit before it has chance to dequeue the last two items.
}

// This code shows broadcasting event with Broadcast()
//
// it simulate a GUI app with a button, we register an arbitrary number of
// functions that will run when button is clicked.
func TestBroadcast(_ *testing.T) {
	type Button struct {
		Clicked *sync.Cond
	}
	button := Button{Clicked: sync.NewCond(&sync.Mutex{})}

	subscribe := func(c *sync.Cond, fn func()) {
		// goroutineRunning ensure click handler goroutine is running. subscribe
		// won't exit until handler gorouine is confirmed to be running.
		var goroutineRunning sync.WaitGroup
		goroutineRunning.Add(1)
		go func() {
			goroutineRunning.Done() // confirm running.
			c.L.Lock()
			defer c.L.Unlock()
			c.Wait() 				// wait for click event.
			fn()     				// handle click event.
		}()
		goroutineRunning.Wait()
	}

	// clickRegistered ensure program won't exit before handler executed.
	var clickRegistered sync.WaitGroup
	clickRegistered.Add(3)
	subscribe(button.Clicked, func() {
		fmt.Println("Maximizing window")
		clickRegistered.Done()
	})
	subscribe(button.Clicked, func() {
		fmt.Println("Displaying annoying dialog box")
		clickRegistered.Done()
	})
	subscribe(button.Clicked, func() {
		fmt.Println("Mouse clicked")
		clickRegistered.Done()
	})

	button.Clicked.Broadcast() // broadcast click event.

	clickRegistered.Wait()
}
