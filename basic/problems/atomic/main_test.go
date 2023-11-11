package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestPlusPlusNotAtomic(_ *testing.T) {
	i := 0

	var wg sync.WaitGroup
	incr := func() {
		defer wg.Done()
		i++
	}

	wg.Add(2)
	go incr()
	go incr()
	wg.Wait()

	fmt.Println(i) // might be "1" !!
}
