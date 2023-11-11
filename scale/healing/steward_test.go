package steward

import (
	"log"
	"os"
	"testing"
	"time"
)

// TestSteward shows how steward monitor and restart a unhealthy ward 'doWork'.
func TestSteward(_ *testing.T) {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	// doWord is a ward that does nothing but waiting to be canceled, it also
	// doesn't send out any pulses.
	doWork := func(done <-chan any, _ time.Duration) <-chan any {
		log.Println("ward: Hello, I'm irresponsible!")
		go func() {
			<-done
			log.Println("ward: I am halting.")
		}()
		return nil
	}
	doWorkWithSteward := newSteward(4*time.Second, doWork)

	done := make(chan any)
	time.AfterFunc(9*time.Second, func() {
		log.Println("main: halting steward and ward.")
		close(done)
	})

	// 4s pulse will be ignored by this ward.
	for range doWorkWithSteward(done, 4*time.Second) {
	}
	log.Println("Done.")
}


