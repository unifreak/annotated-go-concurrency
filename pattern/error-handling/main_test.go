package main

import (
	"fmt"
	"net/http"
	"testing"
)

// TestAntiPattern shows the anti pattern of error handling for concurrent
// goroutines: The goroutine has been given no choice in the matter. It can’t
// simply swallow the error, and so it does the only sensible thing: it prints
// the error and hopes something is paying attention. Don’t put your goroutines
// in this awkward position.
func TestAntiPattern(_ *testing.T) {
	checkStatus := func(done <-chan any, urls ...string) <-chan *http.Response {
		responses := make(chan *http.Response)
		go func() {
			defer close(responses)
			for _, url := range urls {
				resp, err := http.Get(url)
				if err != nil {
					fmt.Println(err)
					continue
				}
				select {
				case <-done:
					return
				case responses<- resp:
				}
			}
		}()
		return responses
	}

	done := make(chan any)
	defer close(done)

	urls := []string{"https://www.bing.com", "https://badhost"}
	for resp := range checkStatus(done, urls...) {
		fmt.Printf("Response: %v\n", resp.Status)
	}
}

// TestEmbedErrorInResult shows how to solve the error handling problem by embed
// errors within Result.
func TestEmbedErrorInResult(_ *testing.T) {
	type Result struct {
		Error    error
		Response *http.Response
	}

	checkStatus := func(done <-chan interface{}, urls ...string) <-chan Result {
		results := make(chan Result)
		go func() {
			defer close(results)

			for _, url := range urls {
				var result Result
				resp, err := http.Get(url)
				result = Result{Error: err, Response: resp}
				select {
				case <-done:
					return
				case results <- result:
				}
			}
		}()
		return results
	}

	done := make(chan interface{})
	defer close(done)

	urls := []string{"https://bing.com", "https://badhost", "a", "b", "c"}

	results := checkStatus(done, urls...)
	for result := range results {
		// Error handled by calling goroutine.
		if result.Error != nil {
			fmt.Printf("error: %v\n", result.Error)
			continue
		}
		fmt.Printf("Response: %v\n", result.Response.Status)
	}

	// Embedding error also enable us to break if too many error occured.
	errCount := 0
	results = checkStatus(done, urls...)
	for result := range results {
		if result.Error != nil {
			fmt.Printf("error: %v\n", result.Error)
			errCount++
			if errCount >= 3 {
				fmt.Println("Too many errors, breaking!")
				break
			}
			continue
		}
		fmt.Printf("Response: %v\n", result.Response.Status)
	}
}

