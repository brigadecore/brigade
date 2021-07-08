package term

import (
	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/gdamore/tcell"
)

const (
	successIcon = "✔"
	failedIcon  = "✖"
	runningIcon = "▶"
	pendingIcon = "⟳"
	unknownIcon = "?"

	successColor = tcell.ColorGreen
	failedColor  = tcell.ColorRed
	runningColor = tcell.ColorYellow
	pendingColor = tcell.ColorWhite
	unknownColor = tcell.ColorGrey

	successedTextColor = "[green]"
	failedTextColor    = "[red]"
	runningTextColor   = "[yellow]"
	pendingTextColor   = "[white]"
	unknownTextColor   = "[grey]"
)

func getColorFromWorkerPhase(phase core.WorkerPhase) tcell.Color {
	switch phase {
	case core.WorkerPhaseSucceeded:
		return successColor
	case core.WorkerPhaseFailed:
		return failedColor
	case core.WorkerPhaseRunning:
		return runningColor
	case core.WorkerPhasePending:
		return pendingColor
	default:
		return unknownColor
	}
}

func getColorFromJobPhase(phase core.JobPhase) tcell.Color {
	switch phase {
	case core.JobPhaseSucceeded:
		return successColor
	case core.JobPhaseFailed:
		return failedColor
	case core.JobPhaseRunning:
		return runningColor
	case core.JobPhasePending:
		return pendingColor
	default:
		return unknownColor
	}
}

func getTextColorFromJobPhase(phase core.JobPhase) string {
	switch phase {
	case core.JobPhaseSucceeded:
		return successedTextColor
	case core.JobPhaseFailed:
		return failedTextColor
	case core.JobPhaseRunning:
		return runningTextColor
	case core.JobPhasePending:
		return pendingTextColor
	default:
		return unknownTextColor
	}
}

func getIconFromWorkerPhase(phase core.WorkerPhase) string {
	switch phase {
	case core.WorkerPhaseSucceeded:
		return successIcon
	case core.WorkerPhaseFailed:
		return failedIcon
	case core.WorkerPhaseRunning:
		return runningIcon
	case core.WorkerPhasePending:
		return pendingIcon
	default:
		return unknownIcon
	}
}

func getIconFromJobPhase(phase core.JobPhase) string {
	switch phase {
	case core.JobPhaseSucceeded:
		return successIcon
	case core.JobPhaseFailed:
		return failedIcon
	case core.JobPhaseRunning:
		return runningIcon
	case core.JobPhasePending:
		return pendingIcon
	default:
		return unknownIcon
	}
}
