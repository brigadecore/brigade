package merge

// this package offers utility methods to merge things like channels

// Channels merges multiple channels into one channel
func Channels(chans ...<-chan struct{}) <-chan struct{} {
	switch len(chans) {
	case 0:
		c := make(chan struct{})
		close(c)
		return c
	case 1:
		return chans[0]
	default:
		m := len(chans) / 2
		return mergeTwoChannels(
			Channels(chans[:m]...),
			Channels(chans[m:]...))
	}
}

func mergeTwoChannels(a, b <-chan struct{}) <-chan struct{} {
	c := make(chan struct{})

	go func() {
		defer close(c)
		for a != nil || b != nil {
			select {
			case v, ok := <-a:
				if !ok {
					a = nil
					continue
				}
				c <- v
			case v, ok := <-b:
				if !ok {
					b = nil
					continue
				}
				c <- v
			}
		}
	}()
	return c
}
