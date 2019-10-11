package page

import (
	"sync"
	"time"

	"github.com/rivo/tview"
)

// Loader knows how to load and autoreload.
type Loader struct {
	app             *tview.Application
	reloadInterval  time.Duration
	canLoad         chan struct{} // channel used to know when we are ready to load.
	stopReloading   chan struct{} // channel to know when to stop reloading.
	autoreloading   bool
	autoreloadingMu sync.Mutex
}

// NewLoader returns a new loader.
func NewLoader(reloadInterval time.Duration, app *tview.Application) *Loader {
	l := &Loader{
		app:            app,
		reloadInterval: reloadInterval,
		canLoad:        make(chan struct{}),
		stopReloading:  make(chan struct{}),
	}
	return l
}

// LoadPage will ensure only one route is loaded at a time. Also ensures
// that the pages that can autoreload do it if it is required.
func (l *Loader) LoadPage(allowAutoreload bool, f func()) {
	if l.getAutoreloading() {
		// If already reloading stop the previous reload and wait.
		l.stopReloading <- struct{}{}
		// Wait until we can load.
		<-l.canLoad
	}

	// We are ready to load this means nothing is autoreloading.
	l.setAutoreloading(false)

	// Load and if we need to start autoreload, start.
	f()
	if allowAutoreload && l.reloadInterval > 0 {
		l.setAutoreloading(true)
		go l.autoreload(f)
	}
}

// autoreload run is the loop that will run and autoreload everything
func (l *Loader) autoreload(f func()) {
	// When finished allow next autoreload.
	defer func() {
		l.canLoad <- struct{}{}
	}()

	t := time.NewTicker(l.reloadInterval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			// Reload page and force draw with the new loaded data...
			f()
			l.app.Draw()
		case <-l.stopReloading:
			return
		}
	}
}

func (l *Loader) getAutoreloading() bool {
	l.autoreloadingMu.Lock()
	defer l.autoreloadingMu.Unlock()
	return l.autoreloading
}

func (l *Loader) setAutoreloading(v bool) {
	l.autoreloadingMu.Lock()
	defer l.autoreloadingMu.Unlock()
	l.autoreloading = v
}
