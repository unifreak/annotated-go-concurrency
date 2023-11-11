package done

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func printGreeting(done <-chan any) error {
	greeting, err := genGreeting(done)
	if err != nil {
		return err
	}
	fmt.Printf("%s world!\n", greeting)
	return nil
}

func genGreeting(done <-chan any) (string, error) {
	switch locale, err := locale(done); {
	case err != nil:
		return "", err
	case locale == "EN/US":
		return "hello", nil
	}
	return "", fmt.Errorf("unsupported locale")
}

func printFarewell(done <-chan any) error {
	farewell, err := genFarewell(done)
	if err != nil {
		return err
	}
	fmt.Printf("%s world!\n", farewell)
	return nil
}

func genFarewell(done <-chan any) (string, error) {
	switch locale, err := locale(done); {
	case err != nil:
		return "", err
	case locale == "EN/US":
		return "goodbye", nil
	}
	return "", fmt.Errorf("unsupported locale")
}

func locale(done <-chan any) (string, error) {
	select {
	case <-done:
		return "", fmt.Errorf("canceled")
	case <-time.After(3 * time.Second):
	}
	return "EN/US", nil
}

// TestDoneChannel show our traditional way of using the done channel for
// cancellation.
//
// Notice the call graph:
//
//			main
//		/			  \
//	printGreeing	printFarewell
//		|				|
//	genGreeting		genFarewell
//		|				|
//	  locale		  locale
func TestDoneChannel(_ *testing.T) {
	var wg sync.WaitGroup
	done := make(chan any)
	defer close(done)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printGreeting(done); err != nil {
			fmt.Printf("%v", err)
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printFarewell(done); err != nil {
			fmt.Printf("%v", err)
			return
		}
	}()

	wg.Wait()
}

