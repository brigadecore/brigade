package page

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/slok/brigadeterm/pkg/controller"
)

const (
	// JobLogPageName is the name that identifies the JobLogPage page.
	JobLogPageName = "joblog"

	// When streaming, check the job status every N interval to know when the job has finished
	// and we need to reload (closing the stream).
	jobStatusCheckInterval = 5 * time.Second
)

const (
	jobInfoFMT = `
%[1]sJob: [white]%[2]s
%[1]sID: [white]%[3]s
%[1]sStarted: [white]%[4]s
%[1]sDuration: [white]%[5]v`
	jobLogUsage = `[yellow](F5) [white]Reload    [yellow](<-/Del) [white]Back    [yellow](ESC) [white]Home    [yellow](Q) [white]Quit`
)

// JobLog is the page where a build job log will be available.
type JobLog struct {
	controller controller.Controller
	router     *Router
	app        *tview.Application

	// page layout.
	layout tview.Primitive

	// components.
	jobsInfo *tview.TextView
	logBox   *tview.TextView
	usage    *tview.TextView

	// stopStreaming is the way we will stop streaming the previous stream on the textview. If we don't stop
	// multiple writes from different goroutines will be writing to the textview.
	stopStreaming chan struct{}
	// canStream is used when we are ready to stream again (previous stream finished).
	canStream chan struct{}

	registerPageOnce sync.Once
}

// NewJobLog returns a new JobLog.
func NewJobLog(controller controller.Controller, app *tview.Application, router *Router) *JobLog {
	j := &JobLog{
		controller: controller,
		router:     router,
		app:        app,
	}
	j.createComponents()
	return j
}

// createComponents will create all the layout components.
func (j *JobLog) createComponents() {
	j.jobsInfo = tview.NewTextView().
		SetDynamicColors(true)
	j.jobsInfo.SetBorder(true).
		SetBorderColor(tcell.ColorYellow)

	j.logBox = tview.NewTextView().
		SetDynamicColors(true).SetChangedFunc(func() {
		j.app.Draw()
	})
	j.logBox.SetBorder(true).
		SetTitle("Log")

	j.usage = tview.NewTextView().
		SetDynamicColors(true)

	// Create the layout.
	j.layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(j.jobsInfo, 0, 1, false).
		AddItem(j.logBox, 0, 5, true).
		AddItem(j.usage, 1, 1, false)
}

// Register satisfies Page interface.
func (j *JobLog) Register(pages *tview.Pages) {
	j.registerPageOnce.Do(func() {
		pages.AddPage(JobLogPageName, j.layout, true, false)
	})
}

// BeforeLoad satisfies Page interface.
func (j *JobLog) BeforeLoad() {
	j.logBox.ScrollToEnd()
}

// Refresh will refresh all the page data.
func (j *JobLog) Refresh(projectID, buildID, jobID string) {
	ctx := j.controller.JobLogPageContext(jobID)

	// Everything seems ok, fill everything.
	j.fill(ctx, projectID, buildID)

	// Set key handlers.
	j.logBox.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		// Reload.
		case tcell.KeyF5:
			j.router.LoadJobLog(projectID, buildID, jobID)
		// Back.
		case tcell.KeyLeft, tcell.KeyDelete, tcell.KeyBackspace, tcell.KeyBackspace2:
			j.router.LoadBuildJobList(projectID, buildID)
		// Home.
		case tcell.KeyEsc:
			j.router.LoadProjectList()
		// Regular keys handling:
		case tcell.KeyRune:
			switch event.Rune() {
			// Reload.
			case 'r', 'R':
				j.router.LoadJobLog(projectID, buildID, jobID)
			// Exit
			case 'q', 'Q':
				j.router.Exit()
			}
		}
		return event
	})
}

func (j *JobLog) fill(ctx *controller.JobLogPageContext, projectID, buildID string) {
	// If not ready to show he logs go back.
	if ctx.Job == nil {
		j.router.LoadBuildJobList(projectID, buildID)
		return
	}

	j.fillUsage()
	j.fillBuildInfo(ctx)
	j.fillLog(ctx, projectID, buildID)
}

func (j *JobLog) fillUsage() {
	j.usage.Clear()
	j.usage.SetText(jobLogUsage)
}

func (j *JobLog) fillBuildInfo(ctx *controller.JobLogPageContext) {
	if ctx.Job == nil { // Safety check.
		return
	}

	color := getColorFromState(ctx.Job.State)
	textColor := getTextColorFromState(ctx.Job.State)
	j.jobsInfo.SetBorderColor(color)

	j.jobsInfo.Clear()
	info := fmt.Sprintf(jobInfoFMT,
		textColor,
		ctx.Job.Name,
		ctx.Job.ID,
		ctx.Job.Started,
		ctx.Job.Ended.Sub(ctx.Job.Started),
	)
	j.jobsInfo.SetText(info)
}

func (j *JobLog) fillLog(ctx *controller.JobLogPageContext, projectID, buildID string) {
	// There are no concurrent fillLog calls, one cli per app run, it's safe to not use mutexes or channels
	// to guarantee accesses.

	// Are we already streaming on background from a previous stream?
	// If yes close the stream and set to initial state.
	if j.stopStreaming != nil {
		close(j.stopStreaming)
		j.stopStreaming = nil
	}

	// Are we ready to stream again? Wait if we have the canStream channel set
	// This channel will be closed when the streaming goroutine that is already
	// streaming has received the stop streaming call.
	if j.canStream != nil {
		<-j.canStream
	}

	// Clean the textview before starting the new stream.
	j.logBox.Clear()

	// Check if we need streaming logic or is just a plain readcloser with all the logs.
	if ctx.Job.State != controller.RunningState {
		j.copyWithAnsiColors(j.logBox, ctx.Log)
		defer ctx.Log.Close()
		return
	}

	// Start streaming in background (will update the textview
	// and the textview will redraw on every change detected, check
	// `SetChangedFunc` on the textview).
	go j.streamLog(ctx, projectID, buildID)
}

func (j *JobLog) streamLog(ctx *controller.JobLogPageContext, projectID, buildID string) {
	// Initialize control channels for the streaming.
	j.stopStreaming = make(chan struct{})
	j.canStream = make(chan struct{})

	// Save the context on goroutine.
	ss := j.stopStreaming
	cs := j.canStream
	l := ctx.Log

	// Close our reader when finished streaming, ignore if error.
	defer l.Close()

	// When finished we are ready to stream again. Only one can stream at a time.
	defer func() {
		close(cs)
		cs = nil
	}()

	// Run a goroutine to check the state of the job on inteval N.
	// If job finished we could reload everything and stop our streaming.
	go func() {
		t := time.NewTicker(jobStatusCheckInterval)
		defer t.Stop()
		for range t.C {
			// Check if another streaming has been started before finishing this
			// and we need to stop checking this job status.
			select {
			case <-ss:
				return
			default:
			}

			// If not running is time to reload everything.
			if !j.controller.JobRunning(ctx.Job.ID) {
				j.Refresh(projectID, buildID, ctx.Job.ID)
				return
			}
		}
	}()

	// Start showing the stream on the textView.
	// Ignore the copy error.
	j.copyWithAnsiColors(j.logBox, readerFunc(func(p []byte) (n int, err error) {
		select {
		// if we don't want to continue reading return 0.
		case <-ss:
			return 0, io.EOF
		default: // Fallback to read.
			return l.Read(p)
		}
	}))
}

// helper type to create a reader from a func.
type readerFunc func(p []byte) (n int, err error)

func (r readerFunc) Read(p []byte) (n int, err error) { return r(p) }

func (j *JobLog) copyWithAnsiColors(w io.Writer, r io.Reader) {
	cw := tview.ANSIWriter(w)
	io.Copy(cw, r)
}
