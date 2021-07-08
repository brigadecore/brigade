package term

import (
	"context"
	"fmt"
	"time"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"k8s.io/apimachinery/pkg/util/duration"
)

const eventPageName = "event"

const (
	pipelineBlockString = `█`
	pipelineStepTotal   = 50
	pipelineColor       = tcell.ColorYellow
	pipelineTitle       = "Pipeline timeline (total duration: %s)"
)

type eventPage struct {
	*page
	jobsPipeline *tview.Table
	jobsList     *tview.Table
	usage        *tview.TextView
}

func newEventPage(apiClient core.APIClient, app *tview.Application, router PageRouter) *eventPage {
	e := &eventPage{
		page: newPage(apiClient, app, router),
	}

	e.jobsPipeline = tview.NewTable().
		SetBordersColor(pipelineColor)
	e.jobsPipeline.
		SetTitle(fmt.Sprintf(pipelineTitle, time.Millisecond*0)).
		SetBorder(true)

	// Create the job layout (jobs + log).
	e.jobsList = tview.NewTable().
		SetSelectable(true, false)
	e.jobsList.
		SetBorder(true).
		SetTitle("Jobs")

	// Usage.
	e.usage = tview.NewTextView().
		SetDynamicColors(true)

	// Create the layout.
	e.page.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(e.jobsPipeline, 0, 2, false).
		AddItem(e.jobsList, 0, 4, true).
		AddItem(e.usage, 1, 1, false)

	return e
}

func (e *eventPage) refresh(eventID string) {
	event, err := e.apiClient.Events().Get(context.TODO(), eventID)
	if err != nil {
		// TODO: Handle this
	}
	e.fill(event)

	// Set key handlers.
	e.jobsList.SetInputCapture(func(evt *tcell.EventKey) *tcell.EventKey {
		switch evt.Key() {
		// Reload.
		case tcell.KeyF5:
			e.router.LoadEventPage(eventID)
		// Back.
		case tcell.KeyLeft, tcell.KeyDelete, tcell.KeyBackspace, tcell.KeyBackspace2:
			e.router.LoadProjectPage(event.ProjectID)
		// Home.
		case tcell.KeyEsc:
			e.router.LoadProjectsPage()
		// Regular keys handling:
		case tcell.KeyRune:
			switch evt.Rune() {
			// Reload.
			case 'r', 'R':
				e.router.LoadEventPage(eventID)
			// Exit
			case 'q', 'Q':
				e.router.Exit()
			}
		}
		return evt
	})

}

func (e *eventPage) fill(event core.Event) {
	e.fillUsage()
	e.fillPipelineTimeline(event)
	e.fillJobsList(event)
}

func (e *eventPage) fillUsage() {
	e.usage.Clear()
	e.usage.SetText("[yellow](F5) [white]Reload    [yellow](<-/Del) [white]Back    [yellow](ESC) [white]Home    [yellow](Q) [white]Quit")
}

func (e *eventPage) fillPipelineTimeline(event core.Event) {
	// Create our pipeline timeline.
	e.jobsPipeline.Clear()

	// Get timing information, first job, last job and the total runtime time of the run jobs.
	first, _, totalDuration := e.getJobTimingData(event)
	stepDuration := int(totalDuration.Nanoseconds()) / pipelineStepTotal

	// If there is no duration then don't fill the pipeline
	if stepDuration == 0 {
		return
	}

	offset := 1                // Ignore the first cell (is the name of the job).
	rowsBetweenMultiplier := 2 // Left a row between job rows.

	// Create one row for each job.
	for i, job := range event.Worker.Jobs {
		// Name of job.
		e.jobsPipeline.SetCell(i*rowsBetweenMultiplier, 0, &tview.TableCell{Text: job.Name, Align: tview.AlignLeft, Color: pipelineColor})

		// Get length of pipeline.
		jobDuration := job.Status.Ended.Sub(*job.Status.Started)
		pipelineLen := int(jobDuration.Nanoseconds()) / stepDuration

		// Get the start point of the job by getting the start point and
		// calculating the diff until the start of the current job.
		startOffsetTime := job.Status.Started.Sub(first)
		startOffset := int(startOffsetTime.Nanoseconds()) / stepDuration

		for j := startOffset; j < startOffset+pipelineLen; j++ {
			e.jobsPipeline.SetCell(i*rowsBetweenMultiplier, offset+j, &tview.TableCell{Text: pipelineBlockString, BackgroundColor: pipelineColor, Color: pipelineColor})
		}
	}

	// Set title name:
	e.jobsPipeline.SetTitle(fmt.Sprintf(pipelineTitle, totalDuration.Truncate(1*time.Second)))
}

func (e *eventPage) getJobTimingData(event core.Event) (first, last time.Time, totalDuration time.Duration) {
	if len(event.Worker.Jobs) < 1 {
		return time.Time{}, time.Time{}, 0
	}
	first = *event.Worker.Jobs[0].Status.Started
	last = *event.Worker.Jobs[0].Status.Ended

	for _, job := range event.Worker.Jobs[1:] {
		// If running is not count.
		if !job.Status.Phase.IsTerminal() {
			continue
		}
		if job.Status.Started.Before(first) {
			first = *job.Status.Started
		}
		if job.Status.Ended.After(last) {
			last = *job.Status.Ended
		}
	}

	return first, last, last.Sub(first)
}

func (e *eventPage) fillJobsList(event core.Event) {
	const (
		statusCol int = iota
		nameCol
		imageCol
		ageCol
		startedCol
		endedCol
		durationCol
	)

	e.jobsList.Clear()

	// Set header.
	e.jobsList.SetCell(0, statusCol, &tview.TableCell{Align: tview.AlignCenter, Color: tcell.ColorYellow})
	e.jobsList.SetCell(0, nameCol, &tview.TableCell{Text: "Name", Align: tview.AlignCenter, Color: tcell.ColorYellow})
	e.jobsList.SetCell(0, imageCol, &tview.TableCell{Text: "Image", Align: tview.AlignCenter, Color: tcell.ColorYellow})
	e.jobsList.SetCell(0, ageCol, &tview.TableCell{Text: "Age", Align: tview.AlignCenter, Color: tcell.ColorYellow})
	e.jobsList.SetCell(0, startedCol, &tview.TableCell{Text: "Started", Align: tview.AlignCenter, Color: tcell.ColorYellow})
	e.jobsList.SetCell(0, endedCol, &tview.TableCell{Text: "Ended", Align: tview.AlignCenter, Color: tcell.ColorYellow})
	e.jobsList.SetCell(0, durationCol, &tview.TableCell{Text: "Duration", Align: tview.AlignCenter, Color: tcell.ColorYellow})

	for r, job := range event.Worker.Jobs {
		row := r + 1
		icon := unknownIcon
		color := unknownColor
		// Select row color and symbol.
		icon = getIconFromJobPhase(job.Status.Phase)
		color = getColorFromJobPhase(job.Status.Phase)
		// Fill table.
		e.jobsList.SetCell(row, statusCol, &tview.TableCell{Text: icon, Align: tview.AlignLeft, Color: color})
		e.jobsList.SetCell(row, nameCol, &tview.TableCell{Text: job.Name, Align: tview.AlignLeft, Color: color})
		e.jobsList.SetCell(row, imageCol, &tview.TableCell{Text: job.Spec.PrimaryContainer.Image, Align: tview.AlignLeft, Color: color})
		// TODO: Add age-- needs Job to track create time
		if job.Status.Started != nil {
			started := time.Since(*job.Status.Started).Truncate(time.Second)
			e.jobsList.SetCell(row, startedCol, &tview.TableCell{Text: duration.ShortHumanDuration(started), Align: tview.AlignLeft, Color: color})
		}
		if job.Status.Ended != nil {
			ended := time.Since(*job.Status.Ended).Truncate(time.Second)
			e.jobsList.SetCell(row, endedCol, &tview.TableCell{Text: duration.ShortHumanDuration(ended), Align: tview.AlignLeft, Color: color})
		}
		if job.Status.Started != nil && job.Status.Ended != nil {
			duration := job.Status.Ended.Sub(*job.Status.Started).Truncate(time.Second)
			e.jobsList.SetCell(row, durationCol, &tview.TableCell{Text: fmt.Sprintf("%v", duration), Align: tview.AlignLeft, Color: color})
		}
	}

	// Set selectable to call our jobs.
	e.jobsList.SetSelectedFunc(func(row, column int) {
		// If the row is the header then don't do anything.
		if row > 0 {
			jobName := e.jobsList.GetCell(row, nameCol).Text
			// Load log page
			e.router.LoadJobPage(event.ID, jobName)
		}
	})
}
