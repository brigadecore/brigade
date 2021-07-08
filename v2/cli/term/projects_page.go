package term

import (
	"context"
	"strings"
	"time"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"k8s.io/apimachinery/pkg/util/duration"
)

const projectsPageName = "projects"

type projectsPage struct {
	*page
	projectListFilter string
	projectsTable     *tview.Table
	usage             *tview.TextView
	filterInputField  *tview.InputField
}

func (p *projectsPage) focusFilterForm() {
	p.filterInputField.SetLabelColor(tcell.ColorYellow)
	p.page.app.SetFocus(p.filterInputField)
}

func newProjectsPage(apiClient core.APIClient, app *tview.Application, router PageRouter) *projectsPage {
	p := &projectsPage{
		page: newPage(apiClient, app, router),
	}

	p.filterInputField = tview.NewInputField().
		SetLabel("Filter: ").
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetLabelColor(tcell.ColorBlack).
		SetDoneFunc(func(key tcell.Key) {
			term := p.filterInputField.GetText()
			if term == "" {
				p.filterInputField.SetLabelColor(tcell.ColorBlack)
			} else {
				p.filterInputField.SetLabelColor(tcell.ColorYellow)
			}
			p.projectListFilter = term
			p.filter()
			p.page.app.SetFocus(p.projectsTable)
		})
	// Set up columns
	p.projectsTable = tview.NewTable().SetSelectable(true, false)
	p.projectsTable.
		SetBorder(true).
		SetTitle("Projects")

	// Usage.
	p.usage = tview.NewTextView().
		SetDynamicColors(true)

	// Create the layout.
	p.page.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(p.projectsTable, 0, 1, true).
		AddItem(p.usage, 1, 1, false).
		AddItem(p.filterInputField, 1, 1, false)

	return p
}

func (p *projectsPage) refresh() {
	projects, err := p.apiClient.Projects().List(context.TODO(), nil, nil)
	if err != nil {
		// TODO: Handle this
	}
	mostRecentEventByProject := map[string]core.Event{}
	for _, project := range projects.Items {
		events, err := p.apiClient.Events().List(
			context.TODO(),
			&core.EventsSelector{
				ProjectID: project.ID,
			},
			&meta.ListOptions{
				Limit: 1,
			},
		)
		if err != nil {
			// TODO: Handle this
		}
		if len(events.Items) > 0 {
			mostRecentEventByProject[project.ID] = events.Items[0]
		}
	}

	p.fill(projects, mostRecentEventByProject)

	// Set key handlers.
	p.projectsTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		// Reload.
		case tcell.KeyF5:
			p.router.LoadProjectsPage()
		// Regular keys handling:
		case tcell.KeyRune:
			switch event.Rune() {
			// filter
			case '/':
				p.focusFilterForm()
			// Reload.
			case 'r', 'R':
				p.router.LoadProjectsPage()
			// Exit
			case 'q', 'Q':
				p.router.Exit()
			}
		}
		return event
	})

}

func (p *projectsPage) fill(projects core.ProjectList, mostRecentEventByProject map[string]core.Event) {
	p.fillUsage()
	p.fillProjectList(projects, mostRecentEventByProject)
}

func (p *projectsPage) fillUsage() {
	p.usage.Clear()
	p.usage.SetText("[yellow](F5) [white]Reload    [yellow](/) [white]Filter    [yellow](Q) [white]Quit")
}

func (p *projectsPage) filter() {
	p.router.LoadProjectsPage()
}

func (p *projectsPage) fillProjectList(projects core.ProjectList, mostRecentEventByProject map[string]core.Event) {
	const (
		statusCol int = iota
		idCol
		descriptionCol
		lastEventTimeCol
	)

	// Clear other widgets.
	p.projectsTable.Clear()

	// Set header.
	p.projectsTable.SetCell(0, statusCol, &tview.TableCell{Align: tview.AlignCenter, Color: tcell.ColorYellow})
	p.projectsTable.SetCell(0, idCol, &tview.TableCell{Text: "ID", Align: tview.AlignCenter, Color: tcell.ColorYellow})
	p.projectsTable.SetCell(0, descriptionCol, &tview.TableCell{Text: "Description", Align: tview.AlignCenter, Color: tcell.ColorYellow})
	p.projectsTable.SetCell(0, lastEventTimeCol, &tview.TableCell{Text: "Last Event", Align: tview.AlignCenter, Color: tcell.ColorYellow})

	projectNameIDIndex := map[string]string{}

	// Set body.
	filter := p.projectListFilter
	for r, project := range projects.Items {
		row := r + 1
		if filter != "" && !strings.Contains(strings.ToLower(project.ID), filter) {
			continue
		}

		var since time.Duration
		color := unknownColor
		icon := unknownIcon

		lastEvent, found := mostRecentEventByProject[project.ID]
		if found {
			color = getColorFromWorkerPhase(lastEvent.Worker.Status.Phase)
			icon = getIconFromWorkerPhase(lastEvent.Worker.Status.Phase)
			// Calculate lastevent data.
			since = time.Since(*lastEvent.Worker.Status.Started).Truncate(time.Second)
		}

		// Set the index so we can get the project ID on selection.
		projectNameIDIndex[project.ID] = project.ID

		p.projectsTable.SetCell(row, statusCol, &tview.TableCell{Text: icon, Align: tview.AlignLeft, Color: color})
		p.projectsTable.SetCell(row, idCol, &tview.TableCell{Text: project.ID, Align: tview.AlignLeft, Color: color})
		p.projectsTable.SetCell(row, descriptionCol, &tview.TableCell{Text: project.Description, Align: tview.AlignLeft, Color: color})
		p.projectsTable.SetCell(row, lastEventTimeCol, &tview.TableCell{Text: duration.ShortHumanDuration(since), Align: tview.AlignLeft, Color: color})
	}

	// Set selectable to call our jobs.
	p.projectsTable.SetSelectedFunc(func(row, column int) {
		// If the row is the header then don't do anything.
		if row > 0 {
			// Get project ID cell and from commit the event ID.
			cell := p.projectsTable.GetCell(row, idCol)
			projectID := projectNameIDIndex[cell.Text]
			// Load event list page.
			p.router.LoadProjectPage(projectID)
		}
	})
}
