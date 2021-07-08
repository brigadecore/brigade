package term

import (
	"context"
	"fmt"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const jobPageName = "job"

const (
	jobInfoFMT = `
%[1]sJob: [white]%[2]s
%[1]sID: [white]%[3]s
%[1]sStarted: [white]%[4]s
%[1]sDuration: [white]%[5]v`
	jobLogUsage = `[yellow](F5) [white]Reload    [yellow](<-/Del) [white]Back    [yellow](ESC) [white]Home    [yellow](Q) [white]Quit`
)

type jobPage struct {
	*page
	// components.
	jobsInfo *tview.TextView
	logBox   *tview.TextView
	usage    *tview.TextView
}

func newJobPage(apiClient core.APIClient, app *tview.Application, router PageRouter) *jobPage {
	j := &jobPage{
		page: newPage(apiClient, app, router),
	}

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
	j.page.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(j.jobsInfo, 0, 1, false).
		AddItem(j.logBox, 0, 5, true).
		AddItem(j.usage, 1, 1, false)

	return j
}

func (j *jobPage) refresh(eventID, jobName string) {
	event, err := j.apiClient.Events().Get(context.TODO(), eventID)
	if err != nil {
		// TODO: Handle this
	}
	job, found := event.Worker.Job(jobName)
	if !found {
		// TODO: Handle this
	}

	// Everything seems ok, fill everything.
	j.fill(eventID, job)

	// Set key handlers.
	j.logBox.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		// Reload.
		case tcell.KeyF5:
			j.router.LoadJobPage(eventID, jobName)
		// Back.
		case tcell.KeyLeft, tcell.KeyDelete, tcell.KeyBackspace, tcell.KeyBackspace2:
			j.router.LoadEventPage(eventID)
		// Home.
		case tcell.KeyEsc:
			j.router.LoadProjectsPage()
		// Regular keys handling:
		case tcell.KeyRune:
			switch event.Rune() {
			// Reload.
			case 'r', 'R':
				j.router.LoadJobPage(eventID, jobName)
			// Exit
			case 'q', 'Q':
				j.router.Exit()
			}
		}
		return event
	})
}

func (j *jobPage) fill(eventID string, job core.Job) {
	j.fillUsage()
	j.fillJobInfo(job)
	j.fillLog(eventID, job.Name)
}

func (j *jobPage) fillUsage() {
	j.usage.Clear()
	j.usage.SetText(jobLogUsage)
}

func (j *jobPage) fillJobInfo(job core.Job) {
	color := getColorFromJobPhase(job.Status.Phase)
	textColor := getTextColorFromJobPhase(job.Status.Phase)
	j.jobsInfo.SetBorderColor(color)

	j.jobsInfo.Clear()
	info := fmt.Sprintf(
		jobInfoFMT,
		textColor,
		job.Name,
		job.Name,
		job.Status.Started,
		job.Status.Ended.Sub(*job.Status.Started),
	)
	j.jobsInfo.SetText(info)
}

func (j *jobPage) fillLog(eventID, jobName string) {
	j.logBox.Clear()
	go j.streamLog(eventID, jobName)
}

func (j *jobPage) streamLog(eventID, jobName string) {
	// // Initialize control channels for the streaming.
	// j.stopStreaming = make(chan struct{})
	// j.canStream = make(chan struct{})

	// // Save the context on goroutine.
	// ss := j.stopStreaming
	// cs := j.canStream
	// l := ctx.Log

	// // Close our reader when finished streaming, ignore if error.
	// defer l.Close()

	// // When finished we are ready to stream again. Only one can stream at a time.
	// defer func() {
	// 	close(cs)
	// 	cs = nil
	// }()

	// // Run a goroutine to check the state of the job on inteval N.
	// // If job finished we could reload everything and stop our streaming.
	// go func() {
	// 	t := time.NewTicker(5 * time.Second)
	// 	defer t.Stop()
	// 	for range t.C {
	// 		// Check if another streaming has been started before finishing this
	// 		// and we need to stop checking this job status.
	// 		select {
	// 		case <-ss:
	// 			return
	// 		default:
	// 		}

	// 		// If not running is time to reload everything.
	// 		if ctx.Job.Phase != core.JobPhaseRunning {
	// 			j.Refresh(projectID, eventID, ctx.Job.Name)
	// 			return
	// 		}
	// 	}
	// }()

	// // Start showing the stream on the textView.
	// // Ignore the copy error.
	// j.copyWithAnsiColors(j.logBox, readerFunc(func(p []byte) (n int, err error) {
	// 	select {
	// 	// if we don't want to continue reading return 0.
	// 	case <-ss:
	// 		return 0, io.EOF
	// 	default: // Fallback to read.
	// 		return l.Read(p)
	// 	}
	// }))
}
