package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/pkg/errors"
)

// event is a custom representation of a Brigade event that matches what the
// API server provides to the git-initializer.
type event struct {
	Project struct {
		Secrets map[string]string `json:"secrets"`
	} `json:"project"`
	Worker struct {
		Git *sdk.GitConfig `json:"git"`
	} `json:"worker"`
}

// getEvent loads an event from the indicated path on the file system.
func getEvent(path string) (event, error) {
	evt := event{}
	eventPath := "/var/event/event.json"
	data, err := ioutil.ReadFile(eventPath)
	if err != nil {
		return evt,
			errors.Wrapf(err, "error reading event from file %q", eventPath)
	}
	return evt, errors.Wrapf(
		json.Unmarshal(data, &evt),
		"error reading event from file %q",
		eventPath,
	)
}
