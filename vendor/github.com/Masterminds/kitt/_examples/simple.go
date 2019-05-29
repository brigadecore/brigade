package main

import (
	"time"

	"github.com/Masterminds/kitt/progress"
)

func main() {
	p := progress.NewIndicator()
	p.Start("Starting")

	time.Sleep(2 * time.Second)
	p.Message("Still going")

	time.Sleep(2 * time.Second)
	p.Done("Done")

	time.Sleep(20 * time.Millisecond)
	println("The end")
}
