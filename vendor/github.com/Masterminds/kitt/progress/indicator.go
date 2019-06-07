package progress

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Indicator struct {
	Interval time.Duration
	Frames   []string
	Writer   io.Writer

	initialMessage string
	done           chan string
	msg            chan string
}

func NewIndicator() *Indicator {
	return &Indicator{
		Interval: time.Second,
		Writer:   os.Stdout,
		Frames: []string{
			".  ",
			".. ",
			"...",
		},
	}
}

func (i *Indicator) Start(message string) {
	i.done = make(chan string)
	i.msg = make(chan string, 100)
	go i.run()
	i.msg <- message
}

func (i *Indicator) Message(message string) {
	i.msg <- message
}

func (i *Indicator) Done(message string) {
	i.done <- message
}

func (i *Indicator) run() {
	tt := time.NewTicker(i.Interval)
	fr := 0
	mm := i.initialMessage
	maxlen := len(mm)

	for {
		select {
		case m := <-i.msg:
			mm = m
		case <-tt.C:
			pad := maxlen - len(mm)
			if pad < 0 {
				maxlen = len(mm)
			} else if pad > 0 {
				mm += strings.Repeat(" ", pad)
			}
			fmt.Fprintf(i.Writer, "%s %s\r", i.Frames[fr], mm)
			fr++
			if fr >= len(i.Frames) {
				fr = 0
			}
		case mm := <-i.done:
			if pad := maxlen + len(i.Frames[fr]) + 2 - len(mm); pad > 0 {
				mm += strings.Repeat(" ", pad)
			}
			fmt.Println(mm)
			tt.Stop()
			close(i.msg)
			close(i.done)
			return
		}
	}
}
