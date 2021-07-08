package term

import (
	"context"
	"sync"
	"time"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/rivo/tview"
)

type PageRouter interface {
	tview.Primitive
	LoadProjectsPage()
	LoadProjectPage(projectID string)
	LoadEventPage(eventID string)
	LoadJobPage(eventID, jobID string)
	Exit()
}

type pageRouter struct {
	*tview.Pages
	projectsPage      *projectsPage
	projectPage       *projectPage
	eventPage         *eventPage
	jobPage           *jobPage
	app               *tview.Application
	cancelRefreshFn   func()
	cancelRefreshFnMu sync.Mutex
}

func NewPageRouter(apiClient core.APIClient, app *tview.Application) PageRouter {
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
	return r
}

func (r *pageRouter) LoadProjectsPage() {
	r.loadPage(projectsPageName, func() {
		r.projectsPage.refresh()
	})
}

func (r *pageRouter) LoadProjectPage(projectID string) {
	r.loadPage(projectPageName, func() {
		r.projectPage.refresh(projectID)
	})
}

func (r *pageRouter) LoadEventPage(eventID string) {
	r.loadPage(eventPageName, func() {
		r.eventPage.refresh(eventID)
	})
}

func (r *pageRouter) LoadJobPage(eventID, jobID string) {
	r.loadPage(jobPageName, func() {
		r.jobPage.refresh(eventID, jobID)
	})
}

func (r *pageRouter) loadPage(pageName string, fn func()) {
	r.cancelRefreshFnMu.Lock()
	defer r.cancelRefreshFnMu.Unlock()
	if r.cancelRefreshFn != nil {
		r.cancelRefreshFn()
	}
	var ctx context.Context
	ctx, r.cancelRefreshFn = context.WithCancel(context.Background())
	r.SwitchToPage(pageName)
	fn()
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for {
			select {
			case <-ticker.C:
				fn()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (r *pageRouter) Exit() {
	r.app.Stop()
}
