// We collect common generator & stages here, they are used in other files.
package stage

import (
	"fmt"
	. "learn/go/concurrency/pattern/ordone"
	"time"
)

// Repeat stage will repeat the values you pass ot it infinitely until you tell
// it to stop.
func Repeat(
	done <-chan interface{},
	values ...interface{},
) <-chan interface{} {
	valueStream := make(chan interface{})
	go func() {
		defer close(valueStream)
		for {
			for _, v := range values {
				select {
				case <-done:
					return
				case valueStream <- v:
				}
			}
		}
	}()
	return valueStream
}

// Take stage take the first num of items off then exit.
func Take(
	done <-chan interface{},
	valueStream <-chan interface{},
	num int,
) <-chan interface{} {
	takeStream := make(chan interface{})
	go func() {
		defer close(takeStream)

		for i := 0; i < num; i++ {
			select {
			case <-done:
				return
			case takeStream <- <-valueStream: // NOTE the <- <-
			}
		}
	}()
	return takeStream
}

// ReapeatFn repleatly generate values by calling fn.
func RepeatFn(
	done <-chan interface{},
	fn func() interface{},
) <-chan interface{} {
	valueStream := make(chan interface{})
	go func() {
		defer close(valueStream)
		for {
			select {
			case <-done:
				return
			case valueStream <- fn():
			}
		}
	}()
	return valueStream
}

// ToString stage assert a value is a string.
//
// When you need to deal in specific types, you can place a stage that performs
// the type assertion for you.
func ToString(
	done <-chan interface{},
	valueStream <-chan interface{},
) <-chan string {
	stringStream := make(chan string)
	go func() {
		defer close(stringStream)
		for v := range valueStream {
			select {
			case <-done:
				return
			// type assertion
			case stringStream <- v.(string):
			}
		}
	}()
	return stringStream
}

func ToInt(done <-chan interface{}, valueStream <-chan interface{}) <-chan int {
	intStream := make(chan int)
	go func() {
		defer close(intStream)
		for v := range valueStream {
			select {
			case <-done:
				return
			case intStream <- v.(int):
			}
		}
	}()
	return intStream
}

// Sleep stage sleep d before every reading from in and sending to out.
//
// Try implement it yourself and you'll see exactly why we need OrDone Pattern.
// With OrDone'a help, we can implement it like OrDoneSleep.
func Sleep1(done <-chan any, in <-chan any, d time.Duration) <-chan any {
	out := make(chan any)
	go func() {
		defer func() {
			fmt.Println("closing(out)")
			close(out)
		}()

		for v := range in {
			select {
			case <-done:
				return
			case out <- v:
			}
		}
	}()
	return out
}

// Sleep stage sleep d before every reading from in and sending to out.
//
// Try implement it yourself and you'll see exactly why we need OrDone Pattern.
// With OrDone'a help, we can implement it like OrDoneSleep.
func Sleep(done <-chan any, in <-chan any, d time.Duration) <-chan any {
	out := make(chan any)
	go func() {
		defer close(out)

		for {
			select {
			case <-done:
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case <-done:
				case out <- v:
					time.Sleep(d)
				}
			}
		}
	}()
	return out
}

// SleepOrDone is same as Sleep but implemented with the help of OrDone.
func SleepOrDone(done <-chan any, in <-chan any, d time.Duration) <-chan any {
	out := make(chan any)
	go func() {
		defer close(out)

		for v := range OrDone(done, in) {
			select {
			case <-done:
				return
			case out <- v:
				time.Sleep(d)
			}
		}
	}()
	return out
}

// Buffer stage buffer n items for in.
func Buffer(done <-chan any, in <-chan any, n int) <-chan any {
	out := make(chan any, n)
	go func() {
		defer close(out)
		for v := range OrDone(done, in) {
			select {
			case <-done:
				return
			case out <- v:
			}
		}
	}()
	return out
}
