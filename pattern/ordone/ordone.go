package ordone

// OrDone implement the or-done channel pattern. c is the channel which we don't
// have control of.
//
// See the Sleep stage in pipeline/stage.go for why we need ordone pattern.
//
// See tee/, bridge/ and stage/ for usage.
func OrDone(done, c <-chan interface{}) <-chan interface{} {
	valStream := make(chan interface{})
	go func() {
		defer close(valStream)

		// We can not use for-range as this:
		//
		// 		for v := range c {
		// 			select {
		// 			case <-done:
		// 				return
		// 			case valStream <-v	// PROBLEMATIC!! can block forever if
		// 								// reader of valStream lose interest.
		// 			}
		// 		}

		for {
			select {
			case <-done:
				return
			// Since we are not using for-range loop, we need do the ok check
			// manually.
			case v, ok := <-c:
				if !ok {
					return
				}

				// For the same reason preventing us from using for-range loop,
				// We need the inner select to receive from done here, because
				// if we only do:
				//
				// 		valStream <- v
				//
				// it will block forever if the reader of valStream lose interest.
				select {
				case valStream <- v:
				case <-done:
					// We can omit the `return` here, becuase the outer loop will
					// then read the closed done an return accordingly.
				}
			}
		}
	}()
	return valStream
}
