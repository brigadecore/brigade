// +build windows

package browser

import (
	"net/url"
	"os/exec"
)

// Open opens a URL with the default browser.
func Open(u string) (err error) {
	_, err = url.ParseRequestURI(u)
	if err != nil {
		return
	}
	_, err = exec.Command("cmd", "/c", "start", u).Output()
	return
}
