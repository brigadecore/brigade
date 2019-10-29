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
	projectInfoFMT = `[yellow]Project: [white]%s
[yellow]URL: [white]%s
[yellow]Namespace: [white]%s`
	projectBuildListUsage = `[yellow](F5) [white]Reload    [yellow](y) [white]Rerun build	[yellow](<-/Del) [white]Back    [yellow](ESC) [white]Home    [yellow](Q) [white]Quit`
)

const (
	// ProjectBuildListPageName is the name that identifies thi ProjectBuildList page.
	ProjectBuildListPageName = "projectbuildlist"
)

// Table columns.
const (
	bldStatusGlyphCol int = iota
	bldEventTypeCol
	bldVersionCol
	bldIDCol
	bldEndCol
	bldDurationCol
)

// ProjectBuildList is the page where a projects build list will be available.
type ProjectBuildList struct {
	controller controller.Controller
	router     *Router
	app        *tview.Application

	// page layout
	layout tview.Primitive

	// components
	projectInfo *tview.TextView
	buildsTable *tview.Table
	usage       *tview.TextView

	registerPageOnce sync.Once
}

// NewProjectBuildList returns a new ProjectBuildList.
func NewProjectBuildList(controller controller.Controller, app *tview.Application, router *Router) *ProjectBuildList {
	p := &ProjectBuildList{
		controller: controller,
		router:     router,
		app:        app,
	}
	p.createComponents()
	return p
}

// createComponents will create all the layout components.
func (p *ProjectBuildList) createComponents() {
	p.projectInfo = tview.NewTextView().
		SetDynamicColors(true)
	p.projectInfo.SetBorder(true).
		SetBorderColor(tcell.ColorYellow)

	p.buildsTable = tview.NewTable().
		SetSelectable(true, false)
	p.buildsTable.
		SetBorder(true).
		SetTitle("Builds")

	p.usage = tview.NewTextView().
		SetDynamicColors(true)

	// Create the layout.
	p.layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(p.projectInfo, 0, 1, false).
		AddItem(p.buildsTable, 0, 6, true).
		AddItem(p.usage, 1, 1, false)

}

// Register satisfies Page interface.
func (p *ProjectBuildList) Register(pages *tview.Pages) {
	p.registerPageOnce.Do(func() {
		pages.AddPage(ProjectBuildListPageName, p.layout, true, false)
	})
}

// BeforeLoad satisfies Page interface.
func (p *ProjectBuildList) BeforeLoad() {
}

// Refresh will refresh all the page data.
func (p *ProjectBuildList) Refresh(projectID string) {
	ctx := p.controller.ProjectBuildListPageContext(projectID)
	// TODO: check error.
	p.fill(projectID, ctx)

	// Set key handlers.
	p.buildsTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		// Reload.
		case tcell.KeyF5:
			p.router.LoadProjectBuildList(projectID)
		// Back.
		case tcell.KeyLeft, tcell.KeyDelete, tcell.KeyEsc, tcell.KeyBackspace, tcell.KeyBackspace2:
			p.router.LoadProjectList()
		// Regular keys handling:
		case tcell.KeyRune:
			switch event.Rune() {
			// Reload.
			case 'r', 'R':
				p.router.LoadProjectBuildList(projectID)
			// Rerun build
			case 'y', 'Y':
				p.rerunBuild()
				p.router.LoadProjectBuildList(projectID)
			// Exit
			case 'q', 'Q':
				p.router.Exit()
			}
		}
		return event
	})

}

func (p *ProjectBuildList) fill(projectID string, ctx *controller.ProjectBuildListPageContext) {
	p.fillProjectInformation(ctx)
	p.fillUsage()
	p.fillBuildList(projectID, ctx)
}

func (p *ProjectBuildList) fillProjectInformation(ctx *controller.ProjectBuildListPageContext) {
	// Fill the project information.
	p.projectInfo.Clear()
	p.projectInfo.SetText(fmt.Sprintf(projectInfoFMT, ctx.ProjectName, ctx.ProjectURL, ctx.ProjectNS))
}

func (p *ProjectBuildList) fillUsage() {
	// Fill usage (not required).
	p.usage.Clear()
	p.usage.SetText(projectBuildListUsage)
}

func (p *ProjectBuildList) fillBuildList(projectID string, ctx *controller.ProjectBuildListPageContext) {
	// Fill the build table.
	p.buildsTable.Clear()

	// Set header.
	p.buildsTable.SetCell(0, bldStatusGlyphCol, &tview.TableCell{Align: tview.AlignCenter, Color: tcell.ColorYellow})
	p.buildsTable.SetCell(0, bldEventTypeCol, &tview.TableCell{Text: "Event type", Align: tview.AlignCenter, Color: tcell.ColorYellow})
	p.buildsTable.SetCell(0, bldVersionCol, &tview.TableCell{Text: "Version", Align: tview.AlignCenter, Color: tcell.ColorYellow})
	p.buildsTable.SetCell(0, bldIDCol, &tview.TableCell{Text: "ID", Align: tview.AlignCenter, Color: tcell.ColorYellow})
	p.buildsTable.SetCell(0, bldEndCol, &tview.TableCell{Text: "End", Align: tview.AlignCenter, Color: tcell.ColorYellow})
	p.buildsTable.SetCell(0, bldDurationCol, &tview.TableCell{Text: "Duration", Align: tview.AlignCenter, Color: tcell.ColorYellow})

	// TODO order by time.
	rowPosition := 1
	for _, build := range ctx.Builds {
		icon := unknownIcon
		color := unknownColor

		if build != nil {
			// Select row color and symbol.
			icon = getIconFromState(build.State)
			color = getColorFromState(build.State)

			// Fill table.
			p.buildsTable.SetCell(rowPosition, bldStatusGlyphCol, &tview.TableCell{Text: icon, Align: tview.AlignLeft, Color: color})
			p.buildsTable.SetCell(rowPosition, bldEventTypeCol, &tview.TableCell{Text: build.EventType, Align: tview.AlignLeft, Color: color})
			p.buildsTable.SetCell(rowPosition, bldVersionCol, &tview.TableCell{Text: build.Version, Align: tview.AlignLeft, Color: color})
			p.buildsTable.SetCell(rowPosition, bldIDCol, &tview.TableCell{Text: build.ID, Align: tview.AlignLeft, Color: color})
			if hasFinished(build.State) {
				timeAgo := time.Since(build.Ended).Truncate(time.Second * 1)
				p.buildsTable.SetCell(rowPosition, bldEndCol, &tview.TableCell{Text: fmt.Sprintf("%v ago", timeAgo), Align: tview.AlignLeft, Color: color})
				duration := build.Ended.Sub(build.Started).Truncate(time.Second * 1)
				p.buildsTable.SetCell(rowPosition, bldDurationCol, &tview.TableCell{Text: fmt.Sprintf("%v", duration), Align: tview.AlignLeft, Color: color})
			}
		}
		rowPosition++
	}

	// Set selectable to call our jobs.
	p.buildsTable.SetSelectedFunc(func(row, column int) {
		// If the row is the header then don't do anything.
		if row > 0 {
			buildID := p.getBuildIDFromTable(row)
			// Load build job list page.
			p.router.LoadBuildJobList(projectID, buildID)
		}
	})
}

func (p *ProjectBuildList) rerunBuild() {
	// Get in which build of the atable are we.
	row, _ := p.buildsTable.GetSelection()
	// If the row is the header then don't do anything.
	if row > 0 {
		// Get build id and rerun the build.
		buildID := p.getBuildIDFromTable(row)
		p.controller.RerunBuild(buildID)
	}
}

func (p *ProjectBuildList) getBuildIDFromTable(row int) string {
	return p.buildsTable.GetCell(row, bldIDCol).Text
}
