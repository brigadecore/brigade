package term

import (
	"context"
	"fmt"
	"time"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"k8s.io/apimachinery/pkg/util/duration"
)

const projectPageName = "project"

type projectPage struct {
	*page
	projectInfo *tview.TextView
	eventsTable *tview.Table
	usage       *tview.TextView
}

func newProjectPage(apiClient core.APIClient, app *tview.Application, router PageRouter) *projectPage {
	p := &projectPage{
		page: newPage(apiClient, app, router),
	}

	p.projectInfo = tview.NewTextView().SetDynamicColors(true)
	p.projectInfo.SetBorder(true).SetBorderColor(tcell.ColorYellow)

	p.eventsTable = tview.NewTable().SetSelectable(true, false)
	p.eventsTable.SetBorder(true).SetTitle("Events")

	p.usage = tview.NewTextView().SetDynamicColors(true)

	// Create the layout.
	p.page.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(p.projectInfo, 0, 1, false).
		AddItem(p.eventsTable, 0, 6, true).
		AddItem(p.usage, 1, 1, false)

	return p
}

func (p *projectPage) refresh(projectID string) {
	project, err := p.apiClient.Projects().Get(context.TODO(), projectID)
	if err != nil {
		// TODO: Handle this
	}
	events, err := p.apiClient.Events().List(
		context.TODO(),
		&core.EventsSelector{
			ProjectID: projectID,
		},
		&meta.ListOptions{},
	)
	if err != nil {
		// TODO: Handle this
	}

	p.fill(project, events)

	// Set key handlers.
	p.eventsTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		// Reload.
		case tcell.KeyF5:
			p.router.LoadProjectPage(projectID)
		// Back.
		case tcell.KeyLeft, tcell.KeyDelete, tcell.KeyEsc, tcell.KeyBackspace, tcell.KeyBackspace2:
			p.router.LoadProjectsPage()
		// Regular keys handling:
		case tcell.KeyRune:
			switch event.Rune() {
			// Reload.
			case 'r', 'R':
				p.router.LoadProjectPage(projectID)
			// Exit
			case 'q', 'Q':
				p.router.Exit()
			}
		}
		return event
	})

}

func (p *projectPage) fill(project core.Project, events core.EventList) {
	p.fillUsage()
	p.fillProjectInformation(project)
	p.fillEventList(events)
}

func (p *projectPage) fillProjectInformation(project core.Project) {
	// Fill the project information.
	p.projectInfo.Clear()
	p.projectInfo.SetText(
		fmt.Sprintf(
			"[yellow]Project: [white]%s\n"+
				"[yellow]Description: [white]%s",
			project.ID,
			project.Description,
		),
	)
}

func (p *projectPage) fillUsage() {
	// Fill usage (not required).
	p.usage.Clear()
	p.usage.SetText("[yellow](F5) [white]Reload    [yellow](y) [white]Retry event	[yellow](<-/Del) [white]Back    [yellow](ESC) [white]Home    [yellow](Q) [white]Quit")
}

func (p *projectPage) fillEventList(events core.EventList) {
	const (
		statusCol int = iota
		idCol
		sourceCol
		typeCol
		ageCol
		startedCol
		endedCol
		durationCol
	)

	// Fill the event table.
	p.eventsTable.Clear()

	// Set header.
	p.eventsTable.SetCell(0, statusCol, &tview.TableCell{Align: tview.AlignCenter, Color: tcell.ColorYellow})
	p.eventsTable.SetCell(0, idCol, &tview.TableCell{Text: "ID", Align: tview.AlignCenter, Color: tcell.ColorYellow})
	p.eventsTable.SetCell(0, sourceCol, &tview.TableCell{Text: "Source", Align: tview.AlignCenter, Color: tcell.ColorYellow})
	p.eventsTable.SetCell(0, typeCol, &tview.TableCell{Text: "Type", Align: tview.AlignCenter, Color: tcell.ColorYellow})
	p.eventsTable.SetCell(0, ageCol, &tview.TableCell{Text: "Age", Align: tview.AlignCenter, Color: tcell.ColorYellow})
	p.eventsTable.SetCell(0, startedCol, &tview.TableCell{Text: "Started", Align: tview.AlignCenter, Color: tcell.ColorYellow})
	p.eventsTable.SetCell(0, endedCol, &tview.TableCell{Text: "Ended", Align: tview.AlignCenter, Color: tcell.ColorYellow})
	p.eventsTable.SetCell(0, durationCol, &tview.TableCell{Text: "Duration", Align: tview.AlignCenter, Color: tcell.ColorYellow})

	for r, event := range events.Items {
		row := r + 1
		icon := unknownIcon
		color := unknownColor
		// Select row color and symbol.
		icon = getIconFromWorkerPhase(event.Worker.Status.Phase)
		color = getColorFromWorkerPhase(event.Worker.Status.Phase)
		// Fill table.
		p.eventsTable.SetCell(row, statusCol, &tview.TableCell{Text: icon, Align: tview.AlignLeft, Color: color})
		p.eventsTable.SetCell(row, idCol, &tview.TableCell{Text: event.ID, Align: tview.AlignLeft, Color: color})
		p.eventsTable.SetCell(row, sourceCol, &tview.TableCell{Text: event.Source, Align: tview.AlignLeft, Color: color})
		p.eventsTable.SetCell(row, typeCol, &tview.TableCell{Text: event.Type, Align: tview.AlignLeft, Color: color})
		age := time.Since(*event.Created).Truncate(time.Second)
		p.eventsTable.SetCell(row, ageCol, &tview.TableCell{Text: duration.ShortHumanDuration(age), Align: tview.AlignLeft, Color: color})
		if event.Worker.Status.Started != nil {
			started := time.Since(*event.Worker.Status.Started).Truncate(time.Second)
			p.eventsTable.SetCell(row, startedCol, &tview.TableCell{Text: duration.ShortHumanDuration(started), Align: tview.AlignLeft, Color: color})
		}
		if event.Worker.Status.Ended != nil {
			ended := time.Since(*event.Worker.Status.Ended).Truncate(time.Second)
			p.eventsTable.SetCell(row, endedCol, &tview.TableCell{Text: duration.ShortHumanDuration(ended), Align: tview.AlignLeft, Color: color})
		}
		if event.Worker.Status.Started != nil && event.Worker.Status.Ended != nil {
			duration := event.Worker.Status.Ended.Sub(*event.Worker.Status.Started).Truncate(time.Second)
			p.eventsTable.SetCell(row, durationCol, &tview.TableCell{Text: fmt.Sprintf("%v", duration), Align: tview.AlignLeft, Color: color})
		}
	}

	// Set selectable to call our jobs.
	p.eventsTable.SetSelectedFunc(func(row, column int) {
		// If the row is the header then don't do anything.
		if row > 0 {
			eventID := p.eventsTable.GetCell(row, idCol).Text
			// Load event job list page.
			p.router.LoadEventPage(eventID)
		}
	})
}
