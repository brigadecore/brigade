package term

import (
	"context"
	"fmt"
	"time"

	"github.com/brigadecore/brigade/sdk/v3/core"
	"github.com/brigadecore/brigade/sdk/v3/meta"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"k8s.io/apimachinery/pkg/util/duration"
)

const projectsPageName = "projects"

// projectsPage is a custom UI component that displays the list of all
// Projects.
type projectsPage struct {
	*page
	projectsContinueValues []string // Stack of "continue" values to aid paging
	projectsTable          *tview.Table
	usage                  *tview.TextView
}

// newProjectsPage returns a custom UI component that displays the list of all
// Projects.
func newProjectsPage(
	apiClient core.APIClient,
	app *tview.Application,
	router *pageRouter,
) *projectsPage {
	p := &projectsPage{
		page:                   newPage(apiClient, app, router),
		projectsContinueValues: []string{""}, // "" == continue value for first page
		projectsTable:          tview.NewTable().SetSelectable(true, false),
		usage:                  tview.NewTextView().SetDynamicColors(true),
	}
	p.projectsTable.SetBorder(true).SetTitle(" Projects ")
	// Create the layout
	p.page.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(
			p.projectsTable,
			0,
			1,    // Proportionate height-- 1 unit
			true, // Bring into focus
		).
		AddItem(
			p.usage, // Menu
			1,       // Fixed height
			0,
			false, // Don't bring into focus
		)
	return p
}

func (p *projectsPage) load(ctx context.Context) {
	p.refresh(ctx)
}

// refresh refreshes the list of all Projects and repaints the page.
func (p *projectsPage) refresh(ctx context.Context) {
	projects, err := p.apiClient.Projects().List(
		ctx,
		nil,
		&meta.ListOptions{
			Continue: p.projectsContinueValues[len(p.projectsContinueValues)-1],
			Limit:    20,
		},
	)
	if err != nil {
		// TODO: Handle this
	}
	mostRecentEventByProject := map[string]core.Event{}
	for _, project := range projects.Items {
		events, err := p.apiClient.Events().List(
			ctx,
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
	p.fillProjectsTable(projects, mostRecentEventByProject)
	p.fillUsage(projects)
	// Key handling...
	p.projectsTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF5: // Reload
			p.router.loadProjectsPage()
		case tcell.KeyRune: // Regular key handling
			switch event.Rune() {
			case 'r', 'R': // Reload
				p.router.loadProjectsPage()
			case 'p', 'P': // Previous page
				// Pop the current page continue value from the stack, but never pop it
				// if it's the only continue value. i.e. Never pop it if it's the empty
				// continue value that gets you the first page.
				if len(p.projectsContinueValues) > 1 {
					p.projectsContinueValues =
						p.projectsContinueValues[:len(p.projectsContinueValues)-1]
					p.router.loadProjectsPage()
				}
			case 'n', 'N': // Next page
				if projects.Continue != "" {
					// Push the continue value for the next page onto the stack
					p.projectsContinueValues =
						append(p.projectsContinueValues, projects.Continue)
					p.router.loadProjectsPage()
				}
			case 'q', 'Q': // Exit
				p.router.exit()
			}
		}
		return event
	})
}

func (p *projectsPage) fillUsage(projects core.ProjectList) {
	usageText := "[yellow](F5 R) [white]Reload"
	if len(p.projectsContinueValues) > 1 {
		usageText = fmt.Sprintf("%s    [yellow](P) [white]Previous Page", usageText)
	}
	if projects.Continue != "" {
		usageText = fmt.Sprintf("%s    [yellow](N) [white]Next Page", usageText)
	}
	usageText = fmt.Sprintf("%s    [yellow](Q) [white]Quit", usageText)
	p.usage.SetText(usageText)
}

func (p *projectsPage) fillProjectsTable(
	projects core.ProjectList,
	mostRecentEventByProject map[string]core.Event,
) {
	const (
		statusCol int = iota
		idCol
		descriptionCol
		lastEventTimeCol
	)
	p.projectsTable.Clear()
	p.projectsTable.SetCell(
		0,
		statusCol,
		&tview.TableCell{
			Align: tview.AlignCenter,
			Color: tcell.ColorYellow,
		},
	).SetCell(
		0,
		idCol,
		&tview.TableCell{
			Text:  "ID",
			Align: tview.AlignCenter,
			Color: tcell.ColorYellow,
		},
	).SetCell(
		0,
		descriptionCol,
		&tview.TableCell{
			Text:  "Description",
			Align: tview.AlignCenter,
			Color: tcell.ColorYellow,
		},
	).SetCell(
		0,
		lastEventTimeCol,
		&tview.TableCell{
			Text:  "Last Event",
			Align: tview.AlignCenter,
			Color: tcell.ColorYellow,
		},
	)
	for r, project := range projects.Items {
		row := r + 1
		var since time.Duration
		color := getColorFromWorkerPhase(core.WorkerPhaseUnknown)
		icon := getIconFromWorkerPhase(core.WorkerPhaseUnknown)
		lastEvent, found := mostRecentEventByProject[project.ID]
		if found {
			color = getColorFromWorkerPhase(lastEvent.Worker.Status.Phase)
			icon = getIconFromWorkerPhase(lastEvent.Worker.Status.Phase)
			if lastEvent.Worker.Status.Started != nil {
				since = time.Since(
					*lastEvent.Worker.Status.Started,
				).Truncate(time.Second)
			}
		}
		p.projectsTable.SetCell(
			row,
			statusCol,
			&tview.TableCell{
				Text:  icon,
				Align: tview.AlignLeft,
				Color: color,
			},
		).SetCell(
			row,
			idCol,
			&tview.TableCell{
				Text:  project.ID,
				Align: tview.AlignLeft,
				Color: color,
			},
		).SetCell(
			row,
			descriptionCol,
			&tview.TableCell{
				Text:  project.Description,
				Align: tview.AlignLeft,
				Color: color,
			},
		).SetCell(
			row,
			lastEventTimeCol,
			&tview.TableCell{
				Text:  duration.ShortHumanDuration(since),
				Align: tview.AlignLeft,
				Color: color,
			},
		)
	}
	p.projectsTable.SetSelectedFunc(func(row, _ int) {
		if row > 0 { // Header row cells aren't selectable
			projectID := p.projectsTable.GetCell(row, idCol).Text
			p.router.loadProjectPage(projectID)
		}
	})
}
