package merge

import (
	"testing"
)

func TestMergeChannels2(t *testing.T) {

	c1 := make(chan struct{})
	c2 := make(chan struct{})

	merged := Channels(c1, c2)
	close(c1)
	close(c2)

	_, ok := <-merged
	if ok {
		t.Fatal("merged chan expected to be closed")
	}
}

func TestMergeChannels3(t *testing.T) {

	c1 := make(chan struct{})
	c2 := make(chan struct{})
	c3 := make(chan struct{})

	merged := Channels(c1, c2, c3)
	close(c1)
	close(c2)
	close(c3)

	_, ok := <-merged
	if ok {
		t.Fatal("merged chan expected to be closed")
	}
}

func TestMergeZero(t *testing.T) {
	merged := Channels()
	_, ok := <-merged
	if ok {
		t.Fatal("merged chan expected to be closed")
	}
}

func TestMergeTwo(t *testing.T) {
	c1 := make(chan struct{})
	c2 := make(chan struct{})
	merged := mergeTwoChannels(c1, c2)
	go func() {
		for i := 0; i < 2; i++ {
			<-merged
		}
	}()
	c1 <- struct{}{}
	c2 <- struct{}{}
	close(c1)
	close(c2)
	_, ok := <-merged
	if ok {
		t.Fatal("merged chan expected to be closed")
	}
}
