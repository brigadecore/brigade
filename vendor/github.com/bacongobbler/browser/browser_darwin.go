// +build darwin

package browser

import (
	"net/url"
	"os/exec"
)

// Open opens a url with the default browser.
func Open(u string) (err error) {
	_, err = url.ParseRequestURI(u)
	if err != nil {
		return
	}
	_, err = exec.Command("open", u).Output()
	return
}
