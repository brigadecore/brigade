package term

import (
	"context"
	"fmt"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const jobPageName = "job"

// jobPage is a custom UI component that displays Job info and a list of
// associated logs.
type jobPage struct {
	*page
	jobInfo         *tview.TextView
	containersTable *tview.Table
	usage           *tview.TextView
}

// newJobPage returns a custom UI component that displays Job info and a list of
// associated logs.
func newJobPage(
	apiClient core.APIClient,
	app *tview.Application,
	router *pageRouter,
) *jobPage {
	j := &jobPage{
		page:            newPage(apiClient, app, router),
		jobInfo:         tview.NewTextView().SetDynamicColors(true),
		containersTable: tview.NewTable().SetSelectable(true, false),
		usage: tview.NewTextView().SetDynamicColors(true).SetText(
			"[yellow](F5) [white]Reload    [yellow](<-/Del) [white]Back    [yellow](L) [white]Logs    [yellow](ESC) [white]Home    [yellow](Q) [white]Quit", // nolint: lll
		),
	}
	j.jobInfo.SetBorder(true).SetBorderColor(tcell.ColorYellow)
	j.containersTable.SetBorder(true).SetTitle(" Containers ")
	// Create the layout
	j.page.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(j.jobInfo, 0, 1, false).
		AddItem(j.containersTable, 0, 5, true).
		AddItem(j.usage, 1, 1, false)
	return j
}

func (j *jobPage) load(ctx context.Context, eventID, jobName string) {
	j.refresh(ctx, eventID, jobName)
}

// refresh refreshes Job info and repaints the page.
func (j *jobPage) refresh(ctx context.Context, eventID, jobName string) {
	event, err := j.apiClient.Events().Get(ctx, eventID)
	if err != nil {
		// TODO: This return is a bandaid fix to stop nil pointer dereference!
		return
	}
	job, found := event.Worker.Job(jobName)
	if !found {
		// TODO: This return is a bandaid fix to stop nil pointer dereference!
		return
	}
	j.fillJobInfo(eventID, job)
	j.fillContainersTable(job)
	// Set key handlers
	j.containersTable.SetInputCapture(
		func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyF5: // Reload
				j.router.loadJobPage(eventID, jobName)
			case // Back
				tcell.KeyLeft,
				tcell.KeyDelete,
				tcell.KeyBackspace,
				tcell.KeyBackspace2:
				j.router.loadEventPage(eventID)
			case tcell.KeyEsc: // Home
				j.router.loadProjectsPage()
			case tcell.KeyRune: // Regular key handling:
				switch event.Rune() {
				case 'r', 'R': // Reload
					j.router.loadJobPage(eventID, jobName)
				case 'l', 'L':
					j.router.loadLogPage(eventID, jobName)
				case 'q', 'Q': // Exit
					j.router.exit()
				}
			}
			return event
		},
	)
}

func (j *jobPage) fillJobInfo(eventID string, job core.Job) {
	j.jobInfo.SetTitle(fmt.Sprintf(" %s: %s ", eventID, job.Name))
	j.jobInfo.SetBorderColor(getColorFromJobPhase(job.Status.Phase))
	j.jobInfo.Clear()
	infoText := fmt.Sprintf(
		`[grey]Primary Image: [white]%s
[grey]Created: [white]%s
[grey]Started: [white]%s
[grey]Ended: [white]%s`,
		job.Spec.PrimaryContainer.Image,
		formatDateTimeToString(job.Created),
		formatDateTimeToString(job.Status.Started),
		formatDateTimeToString(job.Status.Ended),
	)
	if job.Status.Started != nil && job.Status.Ended != nil {
		infoText = fmt.Sprintf(
			"%s\n[grey]Duration: [white]%v",
			infoText,
			job.Status.Ended.Sub(*job.Status.Started),
		)
	}
	j.jobInfo.SetText(infoText)
}

func (j *jobPage) fillContainersTable(job core.Job) {
	const (
		statusCol int = iota
		nameCol
		imageCol
	)

	j.containersTable.Clear()
	j.containersTable.SetCell(
		0,
		statusCol,
		&tview.TableCell{
			Align: tview.AlignCenter,
			Color: tcell.ColorYellow,
		},
	).SetCell(
		0,
		nameCol,
		&tview.TableCell{
			Text:  "Name",
			Align: tview.AlignCenter,
			Color: tcell.ColorYellow,
		},
	).SetCell(
		0,
		imageCol,
		&tview.TableCell{
			Text:  "Image",
			Align: tview.AlignCenter,
			Color: tcell.ColorYellow,
		},
	)

	row := 1
	icon := getIconFromJobPhase(job.Status.Phase)
	color := getColorFromJobPhase(job.Status.Phase)

	j.containersTable.SetCell(
		row,
		statusCol,
		&tview.TableCell{
			Text:  icon,
			Align: tview.AlignLeft,
			Color: color,
		},
	).SetCell(
		row,
		nameCol,
		&tview.TableCell{
			Text:  job.Name,
			Align: tview.AlignLeft,
			Color: color,
		},
	).SetCell(
		row,
		imageCol,
		&tview.TableCell{
			Text:  job.Spec.PrimaryContainer.ContainerSpec.Image,
			Align: tview.AlignLeft,
			Color: color,
		},
	)

	for k, v := range job.Spec.SidecarContainers {
		row++
		j.containersTable.SetCell(
			row,
			nameCol,
			&tview.TableCell{
				Text:  k,
				Align: tview.AlignLeft,
				Color: tcell.ColorWhite,
			},
		).SetCell(
			row,
			imageCol,
			&tview.TableCell{
				Text:  v.Image,
				Align: tview.AlignLeft,
				Color: tcell.ColorWhite,
			},
		)
	}
}
