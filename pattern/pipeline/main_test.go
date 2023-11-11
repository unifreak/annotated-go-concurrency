package main

import (
	"fmt"
	"testing"
)

// TestBatchProcessing shows how to use pipeline for batch processing: the stage
// operate on chunks of data all at once instead of one discrete value at a
// time, like a slice.
func TestBatchProcessing(_ *testing.T) {
	// Notice that for the original data to remain unaltered, each stage has to
	// make a new slice of equal length to store the results of its
	// calculations. That means that the memory footprint of our program at any
	// one time is doubled the size of the slice we send into the start of our
	// pipeline.
	multiply := func(values []int, multiplier int) []int {
		multipliedValues := make([]int, len(values))
		for i, v := range values {
			multipliedValues[i] = v * multiplier
		}
		return multipliedValues
	}

	add := func(values []int, additive int) []int {
		addedValues := make([]int, len(values))
		for i, v := range values {
			addedValues[i] = v + additive
		}
		return addedValues
	}

	ints := []int{1, 2, 3, 4}

	// Combine stages in pipeline.
	for _, v := range add(multiply(ints, 2), 1) {
		fmt.Println(v)
	}

	// It's easy to add additional stage into pipeline.
	fmt.Println()
	for _, v := range multiply(add(multiply(ints, 2), 1), 2) {
		fmt.Println(v)
	}
}

// TestStreamProcessing shows how to use pipeline for stream processing: the
// stage receives and emits one element at a time.
func TestStreamProcessing(_ *testing.T) {
	// Streaming vs. Batch Processing
	//
	// Pro:
	//
	// The memory footprint of our program is back down to only the size of the
	// pipeline’s input.
	//
	// Cons:
	//
	// * We had to pull the pipeline down into the body of the for loop and let
	//   the range do the heavy lifting of feeding our pipeline. Not only does
	//   this limit the reuse of how we feed the pipeline, it also limits our
	//   ability to scale.
	// * We have other problems too. Effectively, we’re instantiating our
	//   pipeline for every iteration of the loop. Though it’s cheap to make
	//   function calls, we’re making three function calls for each iteration
	//   of the loop.
	multiply := func(value, multiplier int) int {
		return value * multiplier
	}

	add := func(value, additive int) int {
		return value + additive
	}

	ints := []int{1, 2, 3, 4}
	for _, v := range ints {
		fmt.Println(multiply(add(multiply(v, 2), 1), 2))
	}
}

// TestChannelPipeline shows how to use channels for constructing pipeline.
func TestChannelPipeline(_ *testing.T) {
	generator := func(done <-chan interface{}, integers ...int) <-chan int {
		intStream := make(chan int)
		go func() {
			defer close(intStream)
			for i := range integers {
				select {
				case <-done:
					return
				case intStream <- i:
				}
			}
		}()
		return intStream
	}

	// The stages are interconnected in two ways: by the common done channel,
	// and by the channels that are passed into subsequent stages of the
	// pipeline.
	multiply := func(
		done <-chan interface{},
		intStream <-chan int,
		multiplier int,
	) <-chan int {
		multipliedStream := make(chan int)
		go func() {
			defer close(multipliedStream)
			for i := range intStream {
				select {
				case <-done:
					return
				case multipliedStream <- i * multiplier:
				}
			}
		}()
		return multipliedStream
	}

	add := func(
		done <-chan interface{},
		intStream <-chan int,
		additive int,
	) <-chan int {
		addedStream := make(chan int)
		go func() {
			defer close(addedStream)
			for i := range intStream {
				select {
				case <-done:
					return
				case addedStream <- i + additive:
				}
			}
		}()
		return addedStream
	}

	// Closing the done channel cascades through the pipeline.
	done := make(chan interface{})
	defer close(done)

	intStream := generator(done, 1, 2, 3, 4)
	pipeline := multiply(done, add(done, multiply(done, intStream, 2), 1), 2)

	// vs. Stream Pipline Using Slice:
	//
	// * We can use a range statement to extract the values.
	// * Inside each stage we can safely execute concurrently because our inputs
	//   and outputs are safe in concurrent contexts.
	// * Also each stage of the pipeline is executing concurrently.
	for v := range pipeline {
		fmt.Println(v)
	}
}

