// +build linux

package browser

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"os/exec"
)

// Open opens a URL with the default browser.
func Open(u string) (err error) {
	_, err = url.ParseRequestURI(u)
	if err != nil {
		return
	}

	// read /proc/version which contains OS version info
	data, err := ioutil.ReadFile("/proc/version")
	if err != nil {
		return fmt.Errorf("cannot read /proc/version: %v", err)
	}

	// version is in format: Linux version 4.4.0-17134-Microsoft
	if bytes.Contains(data, []byte("-Microsoft")) {
		_, err = exec.Command("cmd.exe", "/c", "start", u).Output()
		return
	}

	_, err = exec.Command("xdg-open", u).Output()
	return
}
