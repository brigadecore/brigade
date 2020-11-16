package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
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

	// Get/stream worker status
	router.HandleFunc(
		"/v2/events/{eventID}/worker/status",
		w.AuthFilter.Decorate(w.getOrStreamStatus),
	).Methods(http.MethodGet)

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

func (w *WorkersEndpoints) getOrStreamStatus(
	wr http.ResponseWriter,
	r *http.Request,
) {
	eventID := mux.Vars(r)["eventID"]
	// nolint: errcheck
	watch, _ := strconv.ParseBool(r.URL.Query().Get("watch"))

	if !watch {
		restmachinery.ServeRequest(
			restmachinery.InboundRequest{
				W: wr,
				R: r,
				EndpointLogic: func() (interface{}, error) {
					return w.Service.GetStatus(r.Context(), eventID)
				},
				SuccessCode: http.StatusOK,
			},
		)
		return
	}

	statusCh, err := w.Service.WatchStatus(r.Context(), eventID)
	if err != nil {
		if _, ok := errors.Cause(err).(*meta.ErrNotFound); ok {
			restmachinery.WriteAPIResponse(wr, http.StatusNotFound, errors.Cause(err))
			return
		}
		log.Printf(
			"error retrieving worker status stream for event %q: %s",
			eventID,
			err,
		)
		restmachinery.WriteAPIResponse(
			wr,
			http.StatusInternalServerError,
			&meta.ErrInternalServer{},
		)
		return
	}

	wr.Header().Set("Content-Type", "text/event-stream")
	wr.(http.Flusher).Flush()
	for status := range statusCh {
		statusBytes, err := json.Marshal(status)
		if err != nil {
			log.Println(errors.Wrapf(err, "error marshaling worker status"))
			return
		}
		fmt.Fprint(wr, string(statusBytes))
		wr.(http.Flusher).Flush()
	}
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
