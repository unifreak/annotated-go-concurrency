package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func printGreeting(ctx context.Context) error {
	greeting, err := genGreeting(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("%s world!\n", greeting)
	return nil
}

func printFarewell(ctx context.Context) error {
	farewell, err := genFarewell(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("%s world!\n", farewell)
	return nil
}

func genGreeting(ctx context.Context) (string, error) {
	// Auto cancel after 1s, thereby cancel any children it passes ctx into,
	// namely locale.
	//
	// Note how genGreeting is able to build new ctx without affecting its
	// parent's ctx, this composability enables us to write large systems
	// without mixing concerns throughout call-graph, it's also the reason we
	// defer the cancel() call in the next line (comparing to the traditional
	// done channel implementations in ./done/).
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	switch locale, err := locale(ctx); {
	case err != nil:
		return "", err
	case locale == "EN/US":
		return "hello", nil
	}
	return "", fmt.Errorf("unsupported locale")
}

func genFarewell(ctx context.Context) (string, error) {
	switch locale, err := locale(ctx); {
	case err != nil:
		return "", err
	case locale == "EN/US":
		return "goodbye", nil
	}
	return "", fmt.Errorf("unsupported locale")
}

func locale(ctx context.Context) (string, error) {
	if deadline, ok := ctx.Deadline(); ok {
		// If would exceed deadline, we'd rather fail fast. Although the
		// difference in this iteration of the program is small
		//
		// In programs that may have a high cost for calling the next bit of
		// functionality, this may save a significant amount of time, but at
		// the very least it also allows the function to fail immediately
		// instead of having to wait for the actual timeout to occur.
		//
		// The only catch is that you have to have some idea of how long your
		// subordinate call-graph will take -- an exercise that can be very
		// difficult.
		if deadline.Sub(time.Now().Add(3 * time.Second)) <= 0 {
			return "", context.DeadlineExceeded
		}
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err() // return reason why the ctx is canceled.
	case <-time.After(3 * time.Second):
	}
	return "EN/US", nil
}

// TestSmartLogicWithContext show how context help use to build smart logic like:
//
// * genGreeting only wants to wait one second before abandoning the call to
//   locale.
// * If printGreeting is unsuccessful, we also want to cancel printFarewell.
// * Since locale takes long time to run, we check to see whether it were given
//   a deadline.
func TestSmartLogic(_ *testing.T) {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printGreeting(ctx); err != nil {
			fmt.Printf("cannot print greeting: %v\n", err)
			cancel() // cancel the ctx to notify downward callgraph.
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printFarewell(ctx); err != nil {
			fmt.Printf("cannot print farewell: %v\n", err)
		}
	}()

	wg.Wait()
}

