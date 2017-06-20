/*Package js provides a JavaScript sandbox for Acid.
 */
package js

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/deis/acid/pkg/js/lib"
	"github.com/deis/quokka/pkg/javascript"
	"github.com/deis/quokka/pkg/javascript/libk8s"
)

// defaultScripts is the ordered list of scripts loaded before an eventHandler.
var defaultScripts = []string{
	"js/run.js",
	"js/event.js",
	"js/job.js",
	"js/waitgroup.js",
	"js/runner.js",
}

// HandleEvent creates a default sandbox and then executes the given Acid.js for the given event.
func HandleEvent(e *Event, acidjs []byte) error {
	s, err := New()
	if err != nil {
		return err
	}

	return s.HandleEvent(e, acidjs)
}

// Sandbox gives access to a particular JavaScript runtime that is configured for Acid.
//
// Do not re-use sandboxes.
type Sandbox struct {
	rt      *javascript.Runtime
	scripts []string
}

// New creates a new *Sandbox
func New() (*Sandbox, error) {
	rt := javascript.NewRuntime()
	s := &Sandbox{
		rt: rt,
	}

	s.scripts = make([]string, len(defaultScripts))
	copy(s.scripts, defaultScripts)

	// Add the "built-in" libraries here:
	if err := libk8s.Register(rt.VM()); err != nil {
		return s, err
	}

	// FIXME: This should make its way into quokka.
	rt.VM().Set("sleep", func(seconds int) {
		time.Sleep(time.Duration(seconds) * time.Second)
	})
	return s, nil
}

func (s *Sandbox) HandleEvent(e *Event, script []byte) error {
	event, err := json.Marshal(e)
	if err != nil {
		return err
	}

	// Placeholder for NodeJS compat.
	s.Variable("exports", map[string]interface{}{})

	// Wrap the AcidJS in a function that we can call later.
	acidScript := `var registerEvents = function(events){` + string(script) + `}`
	if err := s.ExecString(acidScript); err != nil {
		return fmt.Errorf("acid.js is not well formed: %s\n%s", err, acidScript)
	}

	for _, p := range s.scripts {
		if err := s.LoadPrecompiled(p); err != nil {
			return fmt.Errorf("%s: %s", p, err)
		}
	}

	// Execute the event.
	return s.ExecString("fireEvent(" + string(event) + ")")
}

// LoadPrecompiled loads scripts that have been precompiled.
//
// The script must reside in the lib.
func (s *Sandbox) LoadPrecompiled(script string) error {
	data, err := lib.Script(script)
	if err != nil {
		return err
	}
	return s.ExecString(data)
}

// Variable Sets a variable in the runtime.
func (s *Sandbox) Variable(name string, val interface{}) {
	s.rt.VM().Set(name, val)
}

// ExecString executes the given string as a JavaScript file.
func (s *Sandbox) ExecString(script string) error {
	_, err := s.rt.VM().Run(script)
	return err
}

// ExecAll takes a list of scripts and executes them.
func (s *Sandbox) ExecAll(scripts ...[]byte) error {
	for _, script := range scripts {
		if _, err := s.rt.VM().Run(script); err != nil {
			return err
		}
	}

	return nil
}
