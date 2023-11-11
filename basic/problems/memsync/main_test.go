// In the following program, we have three critical sections (tagged "1,2,3" in
// comments).
package main

import (
	"fmt"
	"testing"
)

func TestCriticalSection(_ *testing.T) {
	var data int
	go func() { data++ }() 						// 1
	if data == 0 {         						// 2
		fmt.Println("the value is 0.")
	} else {
		fmt.Printf("the value is %v.\n", data) 	// 3
	}
}
