package rest

import (
	"net/http"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/gorilla/mux"
	"github.com/xeipuuv/gojsonschema"
)

// WorkersEndpoints implements restmachinery.Endpoints to provide Worker-related
// URL --> action mappings to a restmachinery.Server.
type WorkersEndpoints struct {
	AuthFilter               restmachinery.Filter
	WorkerStatusSchemaLoader gojsonschema.JSONLoader
	Service                  core.WorkersService
}

// Register is invoked by restmachinery.Server to register Worker-related URL
// --> action mappings to a restmachinery.Server.
func (w *WorkersEndpoints) Register(router *mux.Router) {
	// Start worker
	router.HandleFunc(
		"/v2/events/{eventID}/worker/start",
		w.AuthFilter.Decorate(w.start),
	).Methods(http.MethodPut)

	// Update worker status
	router.HandleFunc(
		"/v2/events/{eventID}/worker/status",
		w.AuthFilter.Decorate(w.updateStatus),
	).Methods(http.MethodPut)

	// Clean up a worker
	router.HandleFunc(
		"/v2/events/{eventID}/worker/cleanup",
		w.AuthFilter.Decorate(w.cleanup),
	).Methods(http.MethodPut)
}

func (w *WorkersEndpoints) start(wr http.ResponseWriter, r *http.Request) {
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: wr,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, w.Service.Start(r.Context(), mux.Vars(r)["eventID"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (w *WorkersEndpoints) updateStatus(
	wr http.ResponseWriter,
	r *http.Request,
) {
	status := core.WorkerStatus{}
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W:                   wr,
			R:                   r,
			ReqBodySchemaLoader: w.WorkerStatusSchemaLoader,
			ReqBodyObj:          &status,
			EndpointLogic: func() (interface{}, error) {
				return nil,
					w.Service.UpdateStatus(r.Context(), mux.Vars(r)["eventID"], status)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (w *WorkersEndpoints) cleanup(
	wr http.ResponseWriter,
	r *http.Request,
) {
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: wr,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil,
					w.Service.Cleanup(r.Context(), mux.Vars(r)["eventID"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}
