package bridge

import (
	"fmt"
	"testing"
)

func TestBridge(_ *testing.T) {
	genVals := func() <-chan <-chan interface{} {
		chanStream := make(chan (<-chan interface{}))
		go func() {
			defer close(chanStream)
			for i := 0; i < 10; i++ {
				stream := make(chan interface{}, 1)
				stream <- i
				close(stream)
				chanStream <- stream
			}
		}()
		return chanStream
	}

	for v := range Bridge(nil, genVals()) {
		fmt.Println(v)
	}
}
