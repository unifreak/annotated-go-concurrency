package steward

import (
	"fmt"
	. "learn/go/concurrency/pattern/pipeline/stage"
	"log"
	"os"
	"testing"
	"time"
)

// TestWardClosure tests steward monitoring a ward returned by doWorkFn.
//
// In the test result, interspersed with the values we receive, we see errors
// from the ward, and our steward detecting them and restarting the ward.
//
// Also notice that we only ever receive values 1 and 2. This is a symptom of
// our ward starting from scratch every time. When developing your wards, if
// our system is sensitive to *duplicate values*, be sure to take that into
// account. We can also consider writing a steward that *exits after a certain
// number of failures*.
//
// In this case, we could have simply made our generator stateful by updating
// the intList we are closed over in every iteration. Such that the valueLoop
// in doWork will be:
//
// 		valueLoop:
// 			for {
// 				intVal := intList[0]
// 				intList = intList[1:]
// 				...
// 			}
func TestWardClosure(_ *testing.T) {
	log.SetFlags(log.Ltime | log.LUTC)
	log.SetOutput(os.Stdout)

	done := make(chan any)
	defer close(done)

	doWork, intStream := doWorkFn(done, 1, 2, -1, 3, 4, 5)
	// Because we expect failures fairly quickly, we'll set the monitoring
	// period at just one millisecond.
	doWorkWithSteward := newSteward(2*time.Millisecond, doWork)
	doWorkWithSteward(done, 1*time.Hour)

	for intVal := range Take(done, intStream, 6) {
		fmt.Printf("Received: %v\n", intVal)
	}
}
