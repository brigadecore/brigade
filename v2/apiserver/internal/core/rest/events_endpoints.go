package rest

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/gorilla/mux"
	"github.com/xeipuuv/gojsonschema"
)

type EventsEndpoints struct {
	AuthFilter              restmachinery.Filter
	EventSchemaLoader       gojsonschema.JSONLoader
	SourceStateSchemaLoader gojsonschema.JSONLoader
	Service                 core.EventsService
}

func (e *EventsEndpoints) Register(router *mux.Router) {
	// Create event
	router.HandleFunc(
		"/v2/events",
		e.AuthFilter.Decorate(e.create),
	).Methods(http.MethodPost)

	// List events
	router.HandleFunc(
		"/v2/events",
		e.AuthFilter.Decorate(e.list),
	).Methods(http.MethodGet)

	// Get event
	router.HandleFunc(
		"/v2/events/{id}",
		e.AuthFilter.Decorate(e.get),
	).Methods(http.MethodGet)

	// Clone event
	router.HandleFunc(
		"/v2/events/{id}/clone",
		e.AuthFilter.Decorate(e.clone),
	).Methods(http.MethodPut)

	// Update event's source state
	router.HandleFunc(
		"/v2/events/{id}/source-state",
		e.AuthFilter.Decorate(e.updateSourceState),
	).Methods(http.MethodPut)

	// Cancel event
	router.HandleFunc(
		"/v2/events/{id}/cancellation",
		e.AuthFilter.Decorate(e.cancel),
	).Methods(http.MethodPut)

	// Cancel a collection of events
	router.HandleFunc(
		"/v2/events/cancellations",
		e.AuthFilter.Decorate(e.cancelMany),
	).Methods(http.MethodPost)

	// Delete event
	router.HandleFunc(
		"/v2/events/{id}",
		e.AuthFilter.Decorate(e.delete),
	).Methods(http.MethodDelete)

	// Delete a collection of events
	router.HandleFunc(
		"/v2/events",
		e.AuthFilter.Decorate(e.deleteMany),
	).Methods(http.MethodDelete)
}

func (e *EventsEndpoints) clone(
	w http.ResponseWriter,
	r *http.Request,
) {
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return e.Service.Clone(
					r.Context(),
					mux.Vars(r)["id"],
				)
			},
			SuccessCode: http.StatusCreated,
		},
	)
}

func (e *EventsEndpoints) create(w http.ResponseWriter, r *http.Request) {
	event := core.Event{}
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: e.EventSchemaLoader,
			ReqBodyObj:          &event,
			EndpointLogic: func() (interface{}, error) {
				return e.Service.Create(r.Context(), event)
			},
			SuccessCode: http.StatusCreated,
		},
	)
}

func (e *EventsEndpoints) list(w http.ResponseWriter, r *http.Request) {
	selector, err := eventsSelectorFromURLQuery(r.URL.Query())
	if err != nil {
		restmachinery.WriteAPIResponse(
			w,
			http.StatusBadRequest,
			err,
		)
	}
	opts := meta.ListOptions{
		Continue: r.URL.Query().Get("continue"),
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		var err error
		if opts.Limit, err = strconv.ParseInt(limitStr, 10, 64); err != nil ||
			opts.Limit < 1 || opts.Limit > 100 {
			restmachinery.WriteAPIResponse(
				w,
				http.StatusBadRequest,
				&meta.ErrBadRequest{
					Reason: fmt.Sprintf(
						`Invalid value %q for "limit" query parameter`,
						limitStr,
					),
				},
			)
			return
		}
	}
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return e.Service.List(r.Context(), selector, opts)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *EventsEndpoints) get(w http.ResponseWriter, r *http.Request) {
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return e.Service.Get(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *EventsEndpoints) updateSourceState(
	w http.ResponseWriter,
	r *http.Request,
) {
	sourceState := core.SourceState{}
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: e.SourceStateSchemaLoader,
			ReqBodyObj:          &sourceState,
			EndpointLogic: func() (interface{}, error) {
				return nil,
					e.Service.UpdateSourceState(
						r.Context(),
						mux.Vars(r)["id"],
						sourceState,
					)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *EventsEndpoints) cancel(w http.ResponseWriter, r *http.Request) {
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.Service.Cancel(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *EventsEndpoints) cancelMany(
	w http.ResponseWriter,
	r *http.Request,
) {
	selector, err := eventsSelectorFromURLQuery(r.URL.Query())
	if err != nil {
		restmachinery.WriteAPIResponse(
			w,
			http.StatusBadRequest,
			err,
		)
	}
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return e.Service.CancelMany(r.Context(), selector)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *EventsEndpoints) delete(w http.ResponseWriter, r *http.Request) {
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.Service.Delete(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *EventsEndpoints) deleteMany(w http.ResponseWriter, r *http.Request) {
	selector, err := eventsSelectorFromURLQuery(r.URL.Query())
	if err != nil {
		restmachinery.WriteAPIResponse(
			w,
			http.StatusBadRequest,
			err,
		)
	}
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return e.Service.DeleteMany(r.Context(), selector)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func eventsSelectorFromURLQuery(
	queryParams url.Values,
) (core.EventsSelector, *meta.ErrBadRequest) {
	selector := core.EventsSelector{}
	if queryParams == nil {
		return selector, nil
	}
	selector.ProjectID = queryParams.Get("projectID")
	selector.Source = queryParams.Get("source")
	qualifiersStr := queryParams.Get("qualifiers")
	if qualifiersStr != "" {
		qualifiersStrs := strings.Split(qualifiersStr, ",")
		selector.Qualifiers = core.Qualifiers{}
		for _, kvStr := range qualifiersStrs {
			kvTokens := strings.SplitN(kvStr, "=", 2)
			if len(kvTokens) != 2 {
				return selector, &meta.ErrBadRequest{
					Reason: fmt.Sprintf(
						`Invalid value %q for "qualifiers" query parameter`,
						qualifiersStr,
					),
				}
			}
			selector.Qualifiers[kvTokens[0]] = kvTokens[1]
		}
	}
	labelsStr := queryParams.Get("labels")
	if labelsStr != "" {
		labelsStrs := strings.Split(labelsStr, ",")
		selector.Labels = map[string]string{}
		for _, kvStr := range labelsStrs {
			kvTokens := strings.SplitN(kvStr, "=", 2)
			if len(kvTokens) != 2 {
				return selector, &meta.ErrBadRequest{
					Reason: fmt.Sprintf(
						`Invalid value %q for "labels" query parameter`,
						labelsStr,
					),
				}
			}
			selector.Labels[kvTokens[0]] = kvTokens[1]
		}
	}
	sourceStateStr := queryParams.Get("sourceState")
	if sourceStateStr != "" {
		sourceStateStrs := strings.Split(sourceStateStr, ",")
		selector.SourceState = map[string]string{}
		for _, kvStr := range sourceStateStrs {
			kvTokens := strings.SplitN(kvStr, "=", 2)
			if len(kvTokens) != 2 {
				return selector, &meta.ErrBadRequest{
					Reason: fmt.Sprintf(
						`Invalid value %q for "sourceState" query parameter`,
						sourceStateStr,
					),
				}
			}
			selector.SourceState[kvTokens[0]] = kvTokens[1]
		}
	}
	selector.Type = queryParams.Get("type")
	workerPhasesStr := queryParams.Get("workerPhases")
	if workerPhasesStr != "" {
		workerPhaseStrs := strings.Split(workerPhasesStr, ",")
		selector.WorkerPhases = make([]core.WorkerPhase, len(workerPhaseStrs))
		for i, workerPhaseStr := range workerPhaseStrs {
			selector.WorkerPhases[i] = core.WorkerPhase(workerPhaseStr)
		}
	}
	return selector, nil
}
