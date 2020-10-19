package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/gorilla/mux"
	"github.com/xeipuuv/gojsonschema"
)

type SecretsEndpoints struct {
	AuthFilter         restmachinery.Filter
	SecretSchemaLoader gojsonschema.JSONLoader
	Service            core.SecretsService
}

func (s *SecretsEndpoints) Register(router *mux.Router) {
	// List Secrets
	router.HandleFunc(
		"/v2/projects/{projectID}/secrets",
		s.AuthFilter.Decorate(s.list),
	).Methods(http.MethodGet)

	// Set Secret
	router.HandleFunc(
		"/v2/projects/{projectID}/secrets/{key}",
		s.AuthFilter.Decorate(s.set),
	).Methods(http.MethodPut)

	// Unset Secret
	router.HandleFunc(
		"/v2/projects/{projectID}/secrets/{key}",
		s.AuthFilter.Decorate(s.unset),
	).Methods(http.MethodDelete)
}

func (s *SecretsEndpoints) list(w http.ResponseWriter, r *http.Request) {
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
				return s.Service.List(r.Context(), mux.Vars(r)["projectID"], opts)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (s *SecretsEndpoints) set(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	secret := core.Secret{}
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: s.SecretSchemaLoader,
			ReqBodyObj:          &secret,
			EndpointLogic: func() (interface{}, error) {
				if key != secret.Key {
					return nil, &meta.ErrBadRequest{
						Reason: "The secret key in the URL path and request body do not " +
							"match.",
					}
				}
				return nil, s.Service.Set(
					r.Context(),
					mux.Vars(r)["projectID"],
					secret,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (s *SecretsEndpoints) unset(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, s.Service.Unset(
					r.Context(),
					mux.Vars(r)["projectID"],
					key,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}
