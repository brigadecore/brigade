package main

import (
	"os"
	"time"

	"github.com/Masterminds/kitt/progress"
)

func main() {
	pi := &progress.Indicator{
		Interval: 200 * time.Millisecond,
		Frames: []string{
			"[*      ]",
			"[.*     ]",
			"[ .*    ]",
			"[  .*   ]",
			"[   .*  ]",
			"[    .* ]",
			"[      *]",
			"[     *.]",
			"[    *. ]",
			"[   *.  ]",
			"[  *.   ]",
			"[ *.    ]",
		},
		Writer: os.Stdout,
	}

	pi.Start("Hello")
	time.Sleep(7 * time.Second)
	pi.Message("Bye")
	time.Sleep(4 * time.Second)
	pi.Done("Done")
	time.Sleep(40 * time.Millisecond)
}
