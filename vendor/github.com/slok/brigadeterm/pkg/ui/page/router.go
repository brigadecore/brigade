package page

import (
	"github.com/rivo/tview"
	"github.com/slok/brigadeterm/pkg/controller"
)

// Router knows how to route the ui from one page to another.
type Router struct {
	pages                *tview.Pages
	projectListPage      *ProjectList
	projectBuildListPage *ProjectBuildList
	buildJobListPage     *BuildJobList
	jobLogPage           *JobLog
	app                  *tview.Application
	loader               *Loader
}

// NewRouter returns a new router.
func NewRouter(app *tview.Application, loader *Loader, controller controller.Controller, pages *tview.Pages) *Router {
	r := &Router{
		pages:  pages,
		app:    app,
		loader: loader,
	}

	// Create the pages.
	r.projectListPage = NewProjectList(controller, app, r)
	r.projectBuildListPage = NewProjectBuildList(controller, app, r)
	r.buildJobListPage = NewBuildJobList(controller, app, r)
	r.jobLogPage = NewJobLog(controller, app, r)

	// Register our pages on the app pages container.
	r.register()

	return r
}

// Register will register the pages on the ui
func (r *Router) register() {
	pages := []Page{
		r.projectListPage,
		r.projectBuildListPage,
		r.buildJobListPage,
		r.jobLogPage,
	}

	// Register all the pages on the ui.
	for _, page := range pages {
		page.Register(r.pages)
	}
}

// LoadProjectList will set the ui on the project list.
func (r *Router) LoadProjectList() {
	r.loader.LoadPage(true, func() {
		r.projectListPage.BeforeLoad()
		r.pages.SwitchToPage(ProjectListPageName)
		r.projectListPage.Refresh()
	})
}

// LoadProjectBuildList will set the ui on the project build list.
func (r *Router) LoadProjectBuildList(projectID string) {
	r.loader.LoadPage(true, func() {
		r.projectBuildListPage.BeforeLoad()
		r.pages.SwitchToPage(ProjectBuildListPageName)
		r.projectBuildListPage.Refresh(projectID)
	})
}

// LoadBuildJobList will set the ui on the build job list.
func (r *Router) LoadBuildJobList(projectID, buildID string) {
	r.loader.LoadPage(true, func() {
		r.buildJobListPage.BeforeLoad()
		r.pages.SwitchToPage(BuildJobListPageName)
		r.buildJobListPage.Refresh(projectID, buildID)
	})
}

// LoadJobLog will set the ui on the build job log.
func (r *Router) LoadJobLog(projectID, buildID, jobID string) {
	r.loader.LoadPage(false, func() {
		r.jobLogPage.BeforeLoad()
		r.pages.SwitchToPage(JobLogPageName)
		r.jobLogPage.Refresh(projectID, buildID, jobID)
	})
}

// Exit will terminate everything.
func (r *Router) Exit() {
	r.app.Stop()
}
