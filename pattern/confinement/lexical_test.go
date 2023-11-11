package confinement

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
)

// TestLexicalConfinedChan shows how lexical confinement works by confining a
// channel. But since channels are already concurrent safe, this is not a very
// interesting example.
//
// (the code is the same as code inside channel:responsibility())
func TestLexicalConfinedChan(_ *testing.T) {
	chanOwner := func() <-chan int {
		result := make(chan int, 5)
		// writing is confined in closure below.
		go func() {
			defer close(result)
			for i := 0; i < 5; i++ {
				result <- i
			}
		}()
		return result
	}

	// type <-chan confines the use of channel to "read only".
	consumer := func(result <-chan int) {
		for i := range result {
			fmt.Printf("Receiving %v\n", i)
		}
		fmt.Println("Done receiving!")
	}

	// reading confined to main goroutine.
	result := chanOwner()
	consumer(result)
}

//  TestLexicalConfinedBytesBuffer shows how to lexically confine a not
//  concurrent-safe data (here is bytes.Buffer).
func TestLexicalConfinedBytesBuffer(_ *testing.T) {
	// closure printData can't see data define afterward, hence confined the
	// later goroutines to only the part of the slice we're passing in.
	printData := func(wg *sync.WaitGroup, data []byte) {
		defer wg.Done()

		var buff bytes.Buffer
		for _, b := range data {
			fmt.Fprintf(&buff, "%c", b)
		}
		fmt.Println(buff.String())
	}

	var wg sync.WaitGroup
	wg.Add(2)
	data := []byte("golang")
	go printData(&wg, data[:3])
	go printData(&wg, data[3:])

	wg.Wait()
}
