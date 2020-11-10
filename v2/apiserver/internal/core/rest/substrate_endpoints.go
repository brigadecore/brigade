package rest

import (
	"net/http"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/gorilla/mux"
)

// SubstrateEndpoints implements restmachinery.Endpoints to provide
// Substrate-related URL --> action mappings to a restmachinery.Server.
type SubstrateEndpoints struct {
	AuthFilter restmachinery.Filter
	Service    core.SubstrateService
}

// Register is invoked by restmachinery.Server to register Project-related
// URL --> action mappings to a restmachinery.Server.
func (s *SubstrateEndpoints) Register(router *mux.Router) {
	// Count running Workers
	router.HandleFunc(
		"/v2/substrate/running-workers",
		s.AuthFilter.Decorate(s.countRunningWorkers),
	).Methods(http.MethodGet)

	// Count running Jobs
	router.HandleFunc(
		"/v2/substrate/running-jobs",
		s.AuthFilter.Decorate(s.countRunningJobs),
	).Methods(http.MethodGet)
}

func (s *SubstrateEndpoints) countRunningWorkers(
	w http.ResponseWriter,
	r *http.Request,
) {
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return s.Service.CountRunningWorkers(r.Context())
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (s *SubstrateEndpoints) countRunningJobs(
	w http.ResponseWriter,
	r *http.Request,
) {
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return s.Service.CountRunningJobs(r.Context())
			},
			SuccessCode: http.StatusOK,
		},
	)
}
