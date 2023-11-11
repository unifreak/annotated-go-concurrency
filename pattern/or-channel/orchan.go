package orchan

// Or implement the Or-Channel Pattern.
func Or(channels ...<-chan interface{}) <-chan interface{} {
	// Recursion base case.
	switch len(channels) {
	case 0:
		return nil
	case 1:
		return channels[0]
	}

	orDone := make(chan interface{})

	// Create a goroutine so we can wait for closing on channels without blocking.
	go func() {
		defer func() {
			close(orDone)
		}()

		switch len(channels) {
		case 2: // This case is for optimization (by trying avoid recursion).
			select {
			case <-channels[0]:
			case <-channels[1]:
			}
		default: // At least 3 channels
			select {
			case <-channels[0]:
			case <-channels[1]:
			case <-channels[2]:
			// Because we don't know the number of channel, we recurse.
			//
			// Also because of how we’re recursing, every recursive call to or
			// will at least have two channels: the third channel plus the
			// orDone we passed by hand.
			//
			// This recurrence relation will destructure the rest of the slice
			// into or-channels to form a tree from which the first signal will
			// return. We also pass in the orDone channel so that when
			// goroutines at the upper of the tree exit, goroutines down the
			// tree also exit.
			case <-Or(append(channels[3:], orDone)...):
			}
		}
	}()
	return orDone
}
