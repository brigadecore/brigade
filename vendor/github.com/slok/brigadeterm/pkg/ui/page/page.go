package page

import (
	"github.com/rivo/tview"
)

// Page is the way we will render the different pages of the ui
type Page interface {
	// Register will register as a page.
	Register(pages *tview.Pages)

	// BeforeLoad will be called when loading the page from one of the other pages.
	BeforeLoad()
}
