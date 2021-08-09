package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/brigadecore/brigade/v2/apiserver/internal/api"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

// JobsEndpoints implements restmachinery.Endpoints to provide Job-related URL
// --> action mappings to a restmachinery.Server.
type JobsEndpoints struct {
	AuthFilter            restmachinery.Filter
	JobSchemaLoader       gojsonschema.JSONLoader
	JobStatusSchemaLoader gojsonschema.JSONLoader
	Service               api.JobsService
}

// Register is invoked by restmachinery.Server to register Job-related URL
// --> action mappings to a restmachinery.Server.
func (j *JobsEndpoints) Register(router *mux.Router) {
	// Create job
	router.HandleFunc(
		"/v2/events/{eventID}/worker/jobs",
		j.AuthFilter.Decorate(j.create),
	).Methods(http.MethodPost)

	// Start job
	router.HandleFunc(
		"/v2/events/{eventID}/worker/jobs/{jobName}/start",
		j.AuthFilter.Decorate(j.start),
	).Methods(http.MethodPut)

	// Get/stream job status
	router.HandleFunc(
		"/v2/events/{eventID}/worker/jobs/{jobName}/status",
		j.AuthFilter.Decorate(j.getOrStreamStatus),
	).Methods(http.MethodGet)

	// Update job status
	router.HandleFunc(
		"/v2/events/{eventID}/worker/jobs/{jobName}/status",
		j.AuthFilter.Decorate(j.updateStatus),
	).Methods(http.MethodPut)

	// Clean up a job
	router.HandleFunc(
		"/v2/events/{eventID}/worker/jobs/{jobName}/cleanup",
		j.AuthFilter.Decorate(j.cleanup),
	).Methods(http.MethodPut)

	// Timeout a job
	router.HandleFunc(
		"/v2/events/{eventID}/worker/jobs/{jobName}/timeout",
		j.AuthFilter.Decorate(j.timeout),
	).Methods(http.MethodPut)
}

func (j *JobsEndpoints) create(w http.ResponseWriter, r *http.Request) {
	job := api.Job{}
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: j.JobSchemaLoader,
			ReqBodyObj:          &job,
			EndpointLogic: func() (interface{}, error) {
				return nil, j.Service.Create(
					r.Context(),
					mux.Vars(r)["eventID"],
					job,
				)
			},
			SuccessCode: http.StatusCreated,
		},
	)
}

func (j *JobsEndpoints) getOrStreamStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := mux.Vars(r)["eventID"]
	jobName := mux.Vars(r)["jobName"]
	// nolint: errcheck
	watch, _ := strconv.ParseBool(r.URL.Query().Get("watch"))

	// Clients can request use of the SSE protocol instead of HTTP/2 streaming.
	// Not every potential client language has equally good support for both of
	// those, so allowing clients to pick is useful.
	sse, _ := strconv.ParseBool(r.URL.Query().Get("sse")) // nolint: errcheck

	if !watch {
		restmachinery.ServeRequest(
			restmachinery.InboundRequest{
				W: w,
				R: r,
				EndpointLogic: func() (interface{}, error) {
					return j.Service.GetStatus(r.Context(), id, jobName)
				},
				SuccessCode: http.StatusOK,
			},
		)
		return
	}

	statusCh, err := j.Service.WatchStatus(r.Context(), id, jobName)
	if err != nil {
		if _, ok := errors.Cause(err).(*meta.ErrNotFound); ok {
			restmachinery.WriteAPIResponse(w, http.StatusNotFound, errors.Cause(err))
			return
		}
		log.Printf(
			"error retrieving job status stream for event %q job %q: %s",
			id,
			jobName,
			err,
		)
		restmachinery.WriteAPIResponse(
			w,
			http.StatusInternalServerError,
			&meta.ErrInternalServer{},
		)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.(http.Flusher).Flush()
	for status := range statusCh {
		statusBytes, err := json.Marshal(status)
		if err != nil {
			log.Println(errors.Wrapf(err, "error marshaling job status"))
			return
		}
		if sse {
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", string(statusBytes))
		} else {
			fmt.Fprint(w, string(statusBytes))
		}
		w.(http.Flusher).Flush()
		if status.Phase.IsTerminal() {
			if sse {
				fmt.Fprintf(w, "event: done\ndata: done\n\n")
				w.(http.Flusher).Flush()
			}
			return
		}
	}
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

func (j *JobsEndpoints) updateStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	status := api.JobStatus{}
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: j.JobStatusSchemaLoader,
			ReqBodyObj:          &status,
			EndpointLogic: func() (interface{}, error) {
				return nil, j.Service.UpdateStatus(
					r.Context(),
					mux.Vars(r)["eventID"],
					mux.Vars(r)["jobName"],
					status,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (j *JobsEndpoints) cleanup(
	w http.ResponseWriter,
	r *http.Request,
) {
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, j.Service.Cleanup(
					r.Context(),
					mux.Vars(r)["eventID"],
					mux.Vars(r)["jobName"],
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (j *JobsEndpoints) timeout(
	w http.ResponseWriter,
	r *http.Request,
) {
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, j.Service.Timeout(
					r.Context(),
					mux.Vars(r)["eventID"],
					mux.Vars(r)["jobName"],
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}
