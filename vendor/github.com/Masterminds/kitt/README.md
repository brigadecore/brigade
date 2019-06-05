# Kitt: Make CLI Pretty

Kitt is a small Go library for improving CLI UIs.

It currently provides the following:

- A configurable progress indicator.

## Usage

Here's an example progress indicator.

```go
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
```

There are all kinds of things you can customize, including the graph
used for the meter itself, as well as the timing.

For more, see the `_examples` directory.
