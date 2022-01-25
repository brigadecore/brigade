package term

import (
	"time"

	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/gdamore/tcell/v2"
)

const (
	textGreen  = "[green]"
	textGrey   = "[grey]"
	textRed    = "[red]"
	textWhite  = "[white]"
	textYellow = "[yellow]"
)

var colorsByWorkerPhase = map[sdk.WorkerPhase]tcell.Color{
	sdk.WorkerPhaseAborted:          tcell.ColorGrey,
	sdk.WorkerPhaseCanceled:         tcell.ColorGrey,
	sdk.WorkerPhaseFailed:           tcell.ColorRed,
	sdk.WorkerPhasePending:          tcell.ColorWhite,
	sdk.WorkerPhaseRunning:          tcell.ColorYellow,
	sdk.WorkerPhaseSchedulingFailed: tcell.ColorRed,
	sdk.WorkerPhaseStarting:         tcell.ColorYellow,
	sdk.WorkerPhaseSucceeded:        tcell.ColorGreen,
	sdk.WorkerPhaseTimedOut:         tcell.ColorRed,
	sdk.WorkerPhaseUnknown:          tcell.ColorGrey,
}

var textColorsByWorkerPhase = map[sdk.WorkerPhase]string{
	sdk.WorkerPhaseAborted:          textGrey,
	sdk.WorkerPhaseCanceled:         textGrey,
	sdk.WorkerPhaseFailed:           textRed,
	sdk.WorkerPhasePending:          textWhite,
	sdk.WorkerPhaseRunning:          textYellow,
	sdk.WorkerPhaseSchedulingFailed: textRed,
	sdk.WorkerPhaseStarting:         textYellow,
	sdk.WorkerPhaseSucceeded:        textGreen,
	sdk.WorkerPhaseTimedOut:         textRed,
	sdk.WorkerPhaseUnknown:          textGrey,
}

var iconsByWorkerPhase = map[sdk.WorkerPhase]string{
	sdk.WorkerPhaseAborted:          "✖",
	sdk.WorkerPhaseCanceled:         "✖",
	sdk.WorkerPhaseFailed:           "✖",
	sdk.WorkerPhasePending:          "⟳",
	sdk.WorkerPhaseRunning:          "▶",
	sdk.WorkerPhaseSchedulingFailed: "✖",
	sdk.WorkerPhaseStarting:         "▶",
	sdk.WorkerPhaseSucceeded:        "✔",
	sdk.WorkerPhaseTimedOut:         "✖",
	sdk.WorkerPhaseUnknown:          "?",
}

var colorsByJobPhase = map[sdk.JobPhase]tcell.Color{
	sdk.JobPhaseAborted:          tcell.ColorGrey,
	sdk.JobPhaseCanceled:         tcell.ColorGrey,
	sdk.JobPhaseFailed:           tcell.ColorRed,
	sdk.JobPhasePending:          tcell.ColorWhite,
	sdk.JobPhaseRunning:          tcell.ColorYellow,
	sdk.JobPhaseSchedulingFailed: tcell.ColorRed,
	sdk.JobPhaseStarting:         tcell.ColorYellow,
	sdk.JobPhaseSucceeded:        tcell.ColorGreen,
	sdk.JobPhaseTimedOut:         tcell.ColorRed,
	sdk.JobPhaseUnknown:          tcell.ColorGrey,
}

var iconsByJobPhase = map[sdk.JobPhase]string{
	sdk.JobPhaseAborted:          "✖",
	sdk.JobPhaseCanceled:         "✖",
	sdk.JobPhaseFailed:           "✖",
	sdk.JobPhasePending:          "⟳",
	sdk.JobPhaseRunning:          "▶",
	sdk.JobPhaseSchedulingFailed: "✖",
	sdk.JobPhaseStarting:         "▶",
	sdk.JobPhaseSucceeded:        "✔",
	sdk.JobPhaseTimedOut:         "✖",
	sdk.JobPhaseUnknown:          "?",
}

func getColorFromWorkerPhase(phase sdk.WorkerPhase) tcell.Color {
	if color, ok := colorsByWorkerPhase[phase]; ok {
		return color
	}
	return tcell.ColorGrey
}

func getTextColorFromWorkerPhase(phase sdk.WorkerPhase) string {
	if color, ok := textColorsByWorkerPhase[phase]; ok {
		return color
	}
	return "[grey]"
}

func getIconFromWorkerPhase(phase sdk.WorkerPhase) string {
	if icon, ok := iconsByWorkerPhase[phase]; ok {
		return icon
	}
	return "?"
}

func getColorFromJobPhase(phase sdk.JobPhase) tcell.Color {
	if color, ok := colorsByJobPhase[phase]; ok {
		return color
	}
	return tcell.ColorGrey
}

func getIconFromJobPhase(phase sdk.JobPhase) string {
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
