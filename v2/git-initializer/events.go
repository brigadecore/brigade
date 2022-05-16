package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/pkg/errors"
)

// event is a git-initializer-specific representation of a Brigade Event.
type event struct {
	Project struct {
		Secrets map[string]string `json:"secrets"`
	} `json:"project"`
	Worker struct {
		Git *sdk.GitConfig `json:"git"`
	} `json:"worker"`
}

// getEvent loads an Event from a JSON file on the file system.
func getEvent() (event, error) {
	evt := event{}
	eventPath := "/var/event/event.json"
	data, err := ioutil.ReadFile(eventPath)
	if err != nil {
		return evt, errors.Wrapf(err, "unable read the event file %q", eventPath)
	}
	err = json.Unmarshal(data, &evt)
	return evt, errors.Wrap(err, "error unmarshaling the event")
}
