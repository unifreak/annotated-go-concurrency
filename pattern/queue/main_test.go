// We use bufio to demo the performance gain of chunking.
package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	. "learn/go/concurrency/pattern/pipeline/stage"
	"log"
	"os"
	"testing"
	"time"
)

// TestPipelineWithoutQueue shows the time efficiency of a pipeline with a slow
// stage, without the help of a queue/buffer.
func TestPipelineWithoutQueue(_ *testing.T) {
	done := make(chan any)
	defer close(done)

	zeros := Take(done, Repeat(done, 0), 3)
	short := Sleep(done, zeros, 1*time.Second)	// the fast stage
	long := Sleep(done, short, 4*time.Second)	// the slow stage
	pipeline := long

	for v := range pipeline {
		fmt.Println(v)
	}
}

// TestPipelineWithQueue show how queue change the pipeline behavior, but doesn't
// help with the total time cost of the pipeline.
func TestPipelineWithQueue(_ *testing.T) {
	done := make(chan any)
	defer close(done)

	zeros := Take(done, Repeat(done, 0), 3)
	short := Sleep(done, zeros, 1*time.Second)	// the fast stage
	buffer := Buffer(done, short, 2)			// a queue with capacity 2
	long := Sleep(done, buffer, 4*time.Second)	// the slow stage
	pipeline := long

	for v := range pipeline {
		fmt.Println(v)
	}
}

func BenchmarkUnbufferedWrite(b *testing.B) {
	performWrite(b, tmpFileOrFatal())
}

// BenchmarkUnbufferedWrite show how "chunking" can help with overall
// performance, when there is a operation that require overhead.
//
// In this example, the overhead is "growing memory" operation, bufio use
// queue/buffer/chunking to decrease the times of this operation have to be
// done.
func BenchmarkBufferedWrite(b *testing.B) {
	bufferredFile := bufio.NewWriter(tmpFileOrFatal())
	performWrite(b, bufio.NewWriter(bufferredFile))
}

func tmpFileOrFatal() *os.File {
	file, err := ioutil.TempFile("", "tmp")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return file
}

func performWrite(b *testing.B, writer io.Writer) {
	done := make(chan interface{})
	defer close(done)

	b.ResetTimer()
	for bt := range Take(done, Repeat(done, byte(0)), b.N) {
		writer.Write([]byte{bt.(byte)})
	}
}
