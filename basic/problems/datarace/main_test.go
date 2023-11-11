package main

import (
	"fmt"
	"testing"
)

func TestDataRace(_ *testing.T) {
	var data int
	go func() {
		data++ // 1
	}()
	if data == 0 { // 2
		fmt.Printf("the value is %v.\n", data) // 3
	}
}

// Possible Results:
//
// output						execution sequence
// ---------------------------------------------------------
// "the value is 0"				2,3,-
// "the value is 1"				2,1,3
// nothing						1,2

// An natual thinking to solve this is to put a Sleep between 1 and 2, but this
// is BAD, it only increase the *possibility* to let the program do what you
// think. Introducing sleeps into your code can be a handy way to debug
// concurrent programs, but they are not a solution.
