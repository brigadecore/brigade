package page

import (
	"github.com/gdamore/tcell"
	"github.com/slok/brigadeterm/pkg/controller"
)

const (
	successedIcon = "✔"
	failedIcon    = "✖"
	runningIcon   = "▶"
	pendingIcon   = "⟳"
	unknownIcon   = "?"

	successedColor = tcell.ColorGreen
	failedColor    = tcell.ColorRed
	runningColor   = tcell.ColorYellow
	pendingColor   = tcell.ColorWhite
	unknownColor   = tcell.ColorGrey

	successedTextColor = "[green]"
	failedTextColor    = "[red]"
	runningTextColor   = "[yellow]"
	pendingTextColor   = "[white]"
	unknownTextColor   = "[grey]"
)

// getColorFromState gets the color from the state.
func getColorFromState(state controller.State) tcell.Color {
	switch state {
	case controller.SuccessedState:
		return successedColor
	case controller.FailedState:
		return failedColor
	case controller.RunningState:
		return runningColor
	case controller.PendingState:
		return pendingColor
	default:
		return unknownColor
	}
}

// getTextFromState gets the text color from the state.
func getTextColorFromState(state controller.State) string {
	switch state {
	case controller.SuccessedState:
		return successedTextColor
	case controller.FailedState:
		return failedTextColor
	case controller.RunningState:
		return runningTextColor
	case controller.PendingState:
		return pendingTextColor
	default:
		return unknownTextColor
	}
}

// getIconFromState gets the icon from the state.
func getIconFromState(state controller.State) string {
	switch state {
	case controller.SuccessedState:
		return successedIcon
	case controller.FailedState:
		return failedIcon
	case controller.RunningState:
		return runningIcon
	case controller.PendingState:
		return pendingIcon
	default:
		return unknownIcon
	}
}

// hasFinished returns true if the state is in one of the finished states.
func hasFinished(state controller.State) bool {
	return state == controller.SuccessedState || state == controller.FailedState
}
