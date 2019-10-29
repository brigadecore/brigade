package page

import (
	"fmt"
	"sync"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"

	"github.com/slok/brigadeterm/pkg/controller"
)

const (
	// BuildJobListPageName is the name that identifies the BuildJobList page.
	BuildJobListPageName = "buildJoblist"
)

const (
	jbStatusGlyphCol int = iota
	jbNameCol
	jbImageCol
	jbIDCol
	jbEndedCol
	jbDurationCol
)

const (
	pipelineBlockString = `█`
	pipelineStepTotal   = 50
	pipelineColor       = tcell.ColorYellow
	pipelineTitle       = "Pipeline timeline (total duration: %s)"

	buildJobListUsage = `[yellow](F5) [white]Reload    [yellow](<-/Del) [white]Back    [yellow](ESC) [white]Home    [yellow](Q) [white]Quit`
)

// BuildJobList is the page where a build job list will be available.
type BuildJobList struct {
	controller controller.Controller
	router     *Router
	app        *tview.Application

	// page layout.
	layout tview.Primitive

	// components.
	jobsPipeline *tview.Table
	jobsList     *tview.Table
	usage        *tview.TextView

	registerPageOnce sync.Once
}

// NewBuildJobList returns a new BuildJobList.
func NewBuildJobList(controller controller.Controller, app *tview.Application, router *Router) *BuildJobList {
	b := &BuildJobList{
		controller: controller,
		router:     router,
		app:        app,
	}
	b.createComponents()
	return b
}

// createComponents will create all the layout components.
func (b *BuildJobList) createComponents() {
	b.jobsPipeline = tview.NewTable().
		SetBordersColor(pipelineColor)
	b.jobsPipeline.
		SetTitle(fmt.Sprintf(pipelineTitle, time.Millisecond*0)).
		SetBorder(true)

	// Create the job layout (jobs + log).
	b.jobsList = tview.NewTable().
		SetSelectable(true, false)
	b.jobsList.
		SetBorder(true).
		SetTitle("Jobs")

	// Usage.
	b.usage = tview.NewTextView().
		SetDynamicColors(true)

	// Create the layout.
	b.layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(b.jobsPipeline, 0, 2, false).
		AddItem(b.jobsList, 0, 4, true).
		AddItem(b.usage, 1, 1, false)
}

// Register satisfies Page interface.
func (b *BuildJobList) Register(pages *tview.Pages) {
	b.registerPageOnce.Do(func() {
		pages.AddPage(BuildJobListPageName, b.layout, true, false)
	})
}

// BeforeLoad satisfies Page interface.
func (b *BuildJobList) BeforeLoad() {
}

// Refresh will refresh all the page data.
func (b *BuildJobList) Refresh(projectID, buildID string) {
	ctx := b.controller.BuildJobListPageContext(buildID)
	// TODO: check error.
	b.fill(projectID, buildID, ctx)

	// Set key handlers.
	b.jobsList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		// Reload.
		case tcell.KeyF5:
			b.router.LoadBuildJobList(projectID, buildID)
		// Back.
		case tcell.KeyLeft, tcell.KeyDelete, tcell.KeyBackspace, tcell.KeyBackspace2:
			b.router.LoadProjectBuildList(projectID)
		// Home.
		case tcell.KeyEsc:
			b.router.LoadProjectList()
		// Regular keys handling:
		case tcell.KeyRune:
			switch event.Rune() {
			// Reload.
			case 'r', 'R':
				b.router.LoadBuildJobList(projectID, buildID)
			// Exit
			case 'q', 'Q':
				b.router.Exit()
			}
		}
		return event
	})

}

func (b *BuildJobList) fill(projectID, buildID string, ctx *controller.BuildJobListPageContext) {
	b.fillUsage()
	b.fillPipelineTimeline(ctx)
	b.fillJobsList(projectID, buildID, ctx)
}

func (b *BuildJobList) fillUsage() {
	b.usage.Clear()
	b.usage.SetText(buildJobListUsage)
}

func (b *BuildJobList) fillPipelineTimeline(ctx *controller.BuildJobListPageContext) {
	// Create our pipeline timeline.
	b.jobsPipeline.Clear()

	// Get timing information, first job, last job and the total runtime time of the run jobs.
	first, _, totalDuration := b.getJobTimingData(ctx)
	stepDuration := int(totalDuration.Nanoseconds()) / pipelineStepTotal

	// If there is no duration then don't fill the pipeline
	if stepDuration == 0 {
		return
	}

	offset := 1                // Ignore the first cell (is the name of the job).
	rowsBetweenMultiplier := 2 // Left a row between job rows.

	// Create one row for each job.
	for i, job := range ctx.Jobs {
		if job == nil {
			continue
		}

		// Name of job.
		b.jobsPipeline.SetCell(i*rowsBetweenMultiplier, 0, &tview.TableCell{Text: job.Name, Align: tview.AlignLeft, Color: pipelineColor})

		// Get length of pipeline.
		jobDuration := job.Ended.Sub(job.Started)
		pipelineLen := int(jobDuration.Nanoseconds()) / stepDuration

		// Get the start point of the job by getting the start point and
		// calculating the diff until the start of the current job.
		startOffsetTime := job.Started.Sub(first)
		startOffset := int(startOffsetTime.Nanoseconds()) / stepDuration

		for j := startOffset; j < startOffset+pipelineLen; j++ {
			b.jobsPipeline.SetCell(i*rowsBetweenMultiplier, offset+j, &tview.TableCell{Text: pipelineBlockString, BackgroundColor: pipelineColor, Color: pipelineColor})
		}
	}

	// Set title name:
	b.jobsPipeline.SetTitle(fmt.Sprintf(pipelineTitle, totalDuration.Truncate(1*time.Second)))
}

func (b *BuildJobList) getJobTimingData(ctx *controller.BuildJobListPageContext) (first, last time.Time, totalDuration time.Duration) {
	if len(ctx.Jobs) < 1 {
		return time.Time{}, time.Time{}, 0
	}
	first = ctx.Jobs[0].Started
	last = ctx.Jobs[0].Ended

	for _, job := range ctx.Jobs[1:] {
		if job == nil {
			continue
		}

		// If running is not count.
		if !hasFinished(job.State) {
			continue
		}
		if job.Started.Before(first) {
			first = job.Started
		}
		if job.Ended.After(last) {
			last = job.Ended
		}
	}

	return first, last, last.Sub(first)
}

func (b *BuildJobList) fillJobsList(projectID, buildID string, ctx *controller.BuildJobListPageContext) {
	b.jobsList.Clear()

	// Set header.
	b.jobsList.SetCell(0, jbStatusGlyphCol, &tview.TableCell{Align: tview.AlignCenter, Color: tcell.ColorYellow})
	b.jobsList.SetCell(0, jbNameCol, &tview.TableCell{Text: "Name", Align: tview.AlignCenter, Color: tcell.ColorYellow})
	b.jobsList.SetCell(0, jbImageCol, &tview.TableCell{Text: "Image", Align: tview.AlignCenter, Color: tcell.ColorYellow})
	b.jobsList.SetCell(0, jbIDCol, &tview.TableCell{Text: "ID", Align: tview.AlignCenter, Color: tcell.ColorYellow})
	b.jobsList.SetCell(0, jbEndedCol, &tview.TableCell{Text: "Ended", Align: tview.AlignCenter, Color: tcell.ColorYellow})
	b.jobsList.SetCell(0, jbDurationCol, &tview.TableCell{Text: "Duration", Align: tview.AlignCenter, Color: tcell.ColorYellow})

	// TODO order by time.
	rowPosition := 1
	for _, job := range ctx.Jobs {

		icon := unknownIcon
		color := unknownColor
		if job != nil {
			// Select row color and symbol.
			icon = getIconFromState(job.State)
			color = getColorFromState(job.State)

			// Fill table.
			b.jobsList.SetCell(rowPosition, jbStatusGlyphCol, &tview.TableCell{Text: icon, Align: tview.AlignLeft, Color: color})
			b.jobsList.SetCell(rowPosition, jbNameCol, &tview.TableCell{Text: job.Name, Align: tview.AlignLeft, Color: color})
			b.jobsList.SetCell(rowPosition, jbImageCol, &tview.TableCell{Text: job.Image, Align: tview.AlignLeft, Color: color})
			b.jobsList.SetCell(rowPosition, jbIDCol, &tview.TableCell{Text: job.ID, Align: tview.AlignLeft, Color: color})
			if hasFinished(job.State) {
				timeAgo := time.Since(job.Ended).Truncate(time.Second * 1)
				b.jobsList.SetCell(rowPosition, jbEndedCol, &tview.TableCell{Text: fmt.Sprintf("%v ago", timeAgo), Align: tview.AlignLeft, Color: color})
				duration := job.Ended.Sub(job.Started).Truncate(time.Second * 1)
				b.jobsList.SetCell(rowPosition, jbDurationCol, &tview.TableCell{Text: fmt.Sprintf("%v", duration), Align: tview.AlignLeft, Color: color})
			}
		}
		rowPosition++
	}

	// Set selectable to call our jobs.
	b.jobsList.SetSelectedFunc(func(row, column int) {
		// If the row is the header then don't do anything.
		if row > 0 {
			jobID := b.jobsList.GetCell(row, jbIDCol).Text
			// Load log page
			b.router.LoadJobLog(projectID, buildID, jobID)
		}
	})
}
