package rest

import (
	"net/http"

	"github.com/brigadecore/brigade/v2/apiserver/internal/api"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/gorilla/mux"
)

type AuthnEndpoints struct {
	AuthFilter restmachinery.Filter
	Service    api.PrincipalsService
}

func (a *AuthnEndpoints) Register(router *mux.Router) {
	// WhoAmI -- information about the currently authenticated principal
	router.HandleFunc(
		"/v2/whoami",
		a.AuthFilter.Decorate(a.whoAmI),
	).Methods(http.MethodGet)
}

func (a *AuthnEndpoints) whoAmI(w http.ResponseWriter, r *http.Request) {
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return a.Service.WhoAmI(r.Context())
			},
			SuccessCode: http.StatusOK,
		},
	)
}
