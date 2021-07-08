package term

import (
	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/rivo/tview"
)

type page struct {
	*tview.Flex
	apiClient core.APIClient
	router    PageRouter
	app       *tview.Application
}

func newPage(apiClient core.APIClient, app *tview.Application, router PageRouter) *page {
	return &page{
		apiClient: apiClient,
		router:    router,
		app:       app,
	}
}
