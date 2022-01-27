package term

import (
	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/rivo/tview"
)

// page is a base for composing custom tview.Pages that are compatible with the
// pageRouter component.
type page struct {
	*tview.Flex               // page behaves like a layout
	apiClient   sdk.APIClient // Used to refresh data
	router      *pageRouter   // Used for routing to other pages on command
	app         *tview.Application
}

// newPage returns a base for composing custom tview.Pages that are compatible
// with the pageRouter component.
func newPage(
	apiClient sdk.APIClient,
	app *tview.Application,
	router *pageRouter,
) *page {
	return &page{
		apiClient: apiClient,
		router:    router,
		app:       app,
	}
}
