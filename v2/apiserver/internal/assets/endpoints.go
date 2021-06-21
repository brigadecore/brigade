package assets

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Endpoints struct{}

func (e *Endpoints) Register(router *mux.Router) {
	router.PathPrefix("/assets/").Handler(
		http.StripPrefix(
			"/assets/",
			http.FileServer(http.Dir("/brigade/assets/")),
		),
	)
}
