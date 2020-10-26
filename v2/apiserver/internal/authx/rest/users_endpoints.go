package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/gorilla/mux"
)

type UsersEndpoints struct {
	AuthFilter restmachinery.Filter
	Service    authx.UsersService
}

func (u *UsersEndpoints) Register(router *mux.Router) {
	// List users
	router.HandleFunc(
		"/v2/users",
		u.AuthFilter.Decorate(u.list),
	).Methods(http.MethodGet)

	// Get user
	router.HandleFunc(
		"/v2/users/{id}",
		u.AuthFilter.Decorate(u.get),
	).Methods(http.MethodGet)

	// Lock user
	router.HandleFunc(
		"/v2/users/{id}/lock",
		u.AuthFilter.Decorate(u.lock),
	).Methods(http.MethodPut)

	// Unlock user
	router.HandleFunc(
		"/v2/users/{id}/lock",
		u.AuthFilter.Decorate(u.unlock),
	).Methods(http.MethodDelete)
}

func (u *UsersEndpoints) list(w http.ResponseWriter, r *http.Request) {
	opts := meta.ListOptions{
		Continue: r.URL.Query().Get("continue"),
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limitStr := r.URL.Query().Get("limit")
		if limitStr != "" {
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
			}
			return
		}
	}
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return u.Service.List(r.Context(), opts)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (u *UsersEndpoints) get(w http.ResponseWriter, r *http.Request) {
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return u.Service.Get(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (u *UsersEndpoints) lock(w http.ResponseWriter, r *http.Request) {
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, u.Service.Lock(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (u *UsersEndpoints) unlock(w http.ResponseWriter, r *http.Request) {
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, u.Service.Unlock(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}
