package term

import (
	"time"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/gdamore/tcell/v2"
)

const (
	textGreen  = "[green]"
	textGrey   = "[grey]"
	textRed    = "[red]"
	textWhite  = "[white]"
	textYellow = "[yellow]"
)

var colorsByWorkerPhase = map[core.WorkerPhase]tcell.Color{
	core.WorkerPhaseAborted:          tcell.ColorGrey,
	core.WorkerPhaseCanceled:         tcell.ColorGrey,
	core.WorkerPhaseFailed:           tcell.ColorRed,
	core.WorkerPhasePending:          tcell.ColorWhite,
	core.WorkerPhaseRunning:          tcell.ColorYellow,
	core.WorkerPhaseSchedulingFailed: tcell.ColorRed,
	core.WorkerPhaseStarting:         tcell.ColorYellow,
	core.WorkerPhaseSucceeded:        tcell.ColorGreen,
	core.WorkerPhaseTimedOut:         tcell.ColorRed,
	core.WorkerPhaseUnknown:          tcell.ColorGrey,
}

var textColorsByWorkerPhase = map[core.WorkerPhase]string{
	core.WorkerPhaseAborted:          textGrey,
	core.WorkerPhaseCanceled:         textGrey,
	core.WorkerPhaseFailed:           textRed,
	core.WorkerPhasePending:          textWhite,
	core.WorkerPhaseRunning:          textYellow,
	core.WorkerPhaseSchedulingFailed: textRed,
	core.WorkerPhaseStarting:         textYellow,
	core.WorkerPhaseSucceeded:        textGreen,
	core.WorkerPhaseTimedOut:         textRed,
	core.WorkerPhaseUnknown:          textGrey,
}

var iconsByWorkerPhase = map[core.WorkerPhase]string{
	core.WorkerPhaseAborted:          "✖",
	core.WorkerPhaseCanceled:         "✖",
	core.WorkerPhaseFailed:           "✖",
	core.WorkerPhasePending:          "⟳",
	core.WorkerPhaseRunning:          "▶",
	core.WorkerPhaseSchedulingFailed: "✖",
	core.WorkerPhaseStarting:         "▶",
	core.WorkerPhaseSucceeded:        "✔",
	core.WorkerPhaseTimedOut:         "✖",
	core.WorkerPhaseUnknown:          "?",
}

var colorsByJobPhase = map[core.JobPhase]tcell.Color{
	core.JobPhaseAborted:          tcell.ColorGrey,
	core.JobPhaseCanceled:         tcell.ColorGrey,
	core.JobPhaseFailed:           tcell.ColorRed,
	core.JobPhasePending:          tcell.ColorWhite,
	core.JobPhaseRunning:          tcell.ColorYellow,
	core.JobPhaseSchedulingFailed: tcell.ColorRed,
	core.JobPhaseStarting:         tcell.ColorYellow,
	core.JobPhaseSucceeded:        tcell.ColorGreen,
	core.JobPhaseTimedOut:         tcell.ColorRed,
	core.JobPhaseUnknown:          tcell.ColorGrey,
}

var iconsByJobPhase = map[core.JobPhase]string{
	core.JobPhaseAborted:          "✖",
	core.JobPhaseCanceled:         "✖",
	core.JobPhaseFailed:           "✖",
	core.JobPhasePending:          "⟳",
	core.JobPhaseRunning:          "▶",
	core.JobPhaseSchedulingFailed: "✖",
	core.JobPhaseStarting:         "▶",
	core.JobPhaseSucceeded:        "✔",
	core.JobPhaseTimedOut:         "✖",
	core.JobPhaseUnknown:          "?",
}

func getColorFromWorkerPhase(phase core.WorkerPhase) tcell.Color {
	if color, ok := colorsByWorkerPhase[phase]; ok {
		return color
	}
	return tcell.ColorGrey
}

func getTextColorFromWorkerPhase(phase core.WorkerPhase) string {
	if color, ok := textColorsByWorkerPhase[phase]; ok {
		return color
	}
	return "[grey]"
}

func getIconFromWorkerPhase(phase core.WorkerPhase) string {
	if icon, ok := iconsByWorkerPhase[phase]; ok {
		return icon
	}
	return "?"
}

func getColorFromJobPhase(phase core.JobPhase) tcell.Color {
	if color, ok := colorsByJobPhase[phase]; ok {
		return color
	}
	return tcell.ColorGrey
}

func getIconFromJobPhase(phase core.JobPhase) string {
	if icon, ok := iconsByJobPhase[phase]; ok {
		return icon
	}
	return "[grey]"
}

// formatDateTimeToString formats a time object to YYYY-MM-DD HH:MM:SS
// and returns it as a string
func formatDateTimeToString(time *time.Time) string {
	if time == nil {
		return ""
	}
	return time.UTC().Format("2006-01-02 15:04:05")
}
