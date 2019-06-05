# Browser

Go lib to open a link with the default web browser in the background. Works on Mac, Linux and Windows.

On Windows and Mac, the `cmd` and `open` utilities are installed by default, so there's no additional requirement necessary. For linux distributions, the `xdg-utils` package is required to open the browser window. `xdg-utils` is required for [Linux Standards Base (LSB)](https://en.wikipedia.org/wiki/Linux_Standard_Base) conformance, so it should be available on most distributions of Linux.

This project is a copypasta of pkg/webbrowser from the [workflow-cli](https://github.com/deis/workflow-cli) project which we worked on at Deis. Chocolates and flowers should be sent to [@Joshua-Anderson](https://github.com/Joshua-Anderson) for all his hard work figuring out the right commands to invoke the default web browser from the terminal.

## Usage

```
package main

import (
	"log"

	"github.com/bacongobbler/browser"
)

func main() {
	if err := browser.Open("https://example.com"); err != nil {
		log.Fatal(err)
	}
}
```

```
$ go get github.com/bacongobbler/browser
$ go run main.go
```
