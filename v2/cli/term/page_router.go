package term

import (
	"context"
	"sync"
	"time"

	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/rivo/tview"
)

// pageRouter is a custom UI component composed of tview.Pages which can be
// refreshed and brought into focus on command.
type pageRouter struct {
	*tview.Pages
	projectsPage      *projectsPage
	projectPage       *projectPage
	eventPage         *eventPage
	jobPage           *jobPage
	logPage           *logPage
	app               *tview.Application
	cancelRefreshFn   func()
	cancelRefreshFnMu sync.Mutex
}

// NewPageRouter returns a custom UI component composed of tview.Pages which
// can be refreshed and brought into focus on command.
func NewPageRouter(
	apiClient sdk.APIClient,
	app *tview.Application,
) tview.Primitive {
	r := &pageRouter{
		Pages: tview.NewPages(),
		app:   app,
	}
	r.projectsPage = newProjectsPage(apiClient, app, r)
	r.AddPage(projectsPageName, r.projectsPage, true, false)
	r.projectPage = newProjectPage(apiClient, app, r)
	r.AddPage(projectPageName, r.projectPage, true, false)
	r.eventPage = newEventPage(apiClient, app, r)
	r.AddPage(eventPageName, r.eventPage, true, false)
	r.jobPage = newJobPage(apiClient, app, r)
	r.AddPage(jobPageName, r.jobPage, true, false)
	r.logPage = newLogPage(apiClient, app, r)
	r.AddPage(logPageName, r.logPage, true, false)
	r.loadProjectsPage()
	return r
}

// loadProjectsPage refreshes the projects page and brings it into focus.
func (r *pageRouter) loadProjectsPage() {
	r.loadPage(
		projectsPageName,
		func(ctx context.Context) {
			r.projectsPage.load(ctx)
		},
		func(ctx context.Context) {
			r.projectsPage.refresh(ctx)
		},
		true,
	)
}

// loadProjectPage refreshes the project page and brings it into focus.
func (r *pageRouter) loadProjectPage(projectID string) {
	r.loadPage(
		projectPageName,
		func(ctx context.Context) {
			r.projectPage.load(ctx, projectID)
		},
		func(ctx context.Context) {
			r.projectPage.refresh(ctx, projectID)
		},
		true,
	)
}

// loadEventPage refreshes the event page and brings it into focus.
func (r *pageRouter) loadEventPage(eventID string) {
	r.loadPage(
		eventPageName,
		func(ctx context.Context) {
			r.eventPage.load(ctx, eventID)
		},
		func(ctx context.Context) {
			r.eventPage.refresh(ctx, eventID)
		},
		true,
	)
}

// loadJobPage refreshes the job page and brings it into focus.
func (r *pageRouter) loadJobPage(eventID, jobID string) {
	r.loadPage(
		jobPageName,
		func(ctx context.Context) {
			r.jobPage.load(ctx, eventID, jobID)
		},
		func(ctx context.Context) {
			r.jobPage.refresh(ctx, eventID, jobID)
		},
		true,
	)
}

// loadLogPage loads a floating window that displays logs and brings it into
// focus.
func (r *pageRouter) loadLogPage(eventID, jobID string) {
	r.loadPage(
		logPageName,
		func(ctx context.Context) {
			r.logPage.load(ctx, eventID, jobID)
		},
		func(ctx context.Context) {
			r.logPage.refresh(ctx, eventID, jobID)
		},
		false,
	)
}

// loadPage can refresh any page and bring it into focus, given the name of the
// page and a refresh function.
func (r *pageRouter) loadPage(
	pageName string,
	loadFn func(context.Context),
	refreshFn func(context.Context),
	hideOthers bool,
) {
	// This is a critical section of code. We only want one page auto-refreshing
	// at a time.
	r.cancelRefreshFnMu.Lock()
	defer r.cancelRefreshFnMu.Unlock()
	// If any page is already auto-refreshing, stop it
	if r.cancelRefreshFn != nil {
		r.cancelRefreshFn()
	}
	// Build a new context for the auto-refresh goroutine to use
	var ctx context.Context
	ctx, r.cancelRefreshFn = context.WithCancel(context.Background())
	if hideOthers {
		r.SwitchToPage(pageName) // Focus page and hide background
	} else {
		r.ShowPage(pageName) // Focus page and keep background
	}
	loadFn(ctx) // Synchronously load the page once
	go func() { // Start auto-refreshing
		ticker := time.NewTicker(2 * time.Second)
		for {
			select {
			case <-ticker.C:
				refreshFn(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// exit stops the associated tview.Application.
func (r *pageRouter) exit() {
	r.app.Stop()
}
