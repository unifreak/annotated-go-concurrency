package orchan

import (
	"testing"
	"time"
)

// TestOrChannel shows how to use Or func.
func TestOrChannel(t *testing.T) {
	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	// How the recursion tree works:
	//
	//		recursion tree					execution sequence marked by number
	//		-----------------------------   ------------------------------------
	//		[recursion@1]
	//		defer close(orDone@1)			2. close(orDone@1)
	//		select:
	//			<-2h
	//			<-5m
	//			<-1s						1. timeup first, close(1s).
	//
	//			[recursion@2]
	//			defer close(orDone@2)		4. close(orDone@2), like closing
	//			<-select:					   orDone@1 at step 2, will signal
	//			  <-1h						   goroutines down to finish (in this
	//			  <-1m						   case there is none)
	//			  <-orDone@1				3. step 2 will unblock here, allow
	//			  							   recursion@2's goroutine to finish
	start := time.Now()
	<-Or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)
	t.Logf("done after %v\n", time.Since(start))
}
