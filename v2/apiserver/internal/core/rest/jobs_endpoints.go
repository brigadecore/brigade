package rest

import (
	"net/http"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/gorilla/mux"
)

// JobsEndpoints implements restmachinery.Endpoints to provide Job-related URL
// --> action mappings to a restmachinery.Server.
type JobsEndpoints struct {
	AuthFilter restmachinery.Filter
	Service    core.JobsService
}

// Register is invoked by restmachinery.Server to register Job-related URL
// --> action mappings to a restmachinery.Server.
func (j *JobsEndpoints) Register(router *mux.Router) {
	// Start job
	router.HandleFunc(
		"/v2/events/{eventID}/worker/jobs/{jobName}/start",
		j.AuthFilter.Decorate(j.start),
	).Methods(http.MethodPut)
}

func (j *JobsEndpoints) start(w http.ResponseWriter, r *http.Request) {
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, j.Service.Start(
					r.Context(),
					mux.Vars(r)["eventID"],
					mux.Vars(r)["jobName"],
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}
