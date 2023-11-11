package confinement

import (
	"fmt"
	"testing"
)

// TestAdHoc shows how adhoc confinement works by convention: data slice is
// available from both the loopData func and the loop over then handleData
// chan, but by convention we're only accessing it from the loopData func.
func TestAdHoc(_ *testing.T) {
	data := make([]int, 4)

	loopData := func(handleData chan<- int) {
		defer close(handleData)
		for i := range data {
			handleData <- data[i]
		}
	}

	handleData := make(chan int)
	go loopData(handleData)

	for num := range handleData {
		fmt.Println(num)
	}
}
