package rest

import (
	"net/http"
	"strconv"

	"github.com/brigadecore/brigade/v2/apiserver/internal/api"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// SessionsEndpoints implements restmachinery.Endpoints to provide
// Session-related URL --> action mappings to a restmachinery.Server.
type SessionsEndpoints struct {
	AuthFilter restmachinery.Filter
	Service    api.SessionsService
}

// Register is invoked by restmachinery.Server to register Session-related
// URL --> action mappings to a restmachinery.Server.
func (s *SessionsEndpoints) Register(router *mux.Router) {
	// Create session
	router.HandleFunc(
		"/v2/sessions",
		s.create, // No filters applied to this request
	).Methods(http.MethodPost)

	// Delete session
	router.HandleFunc(
		"/v2/session",
		s.AuthFilter.Decorate(s.delete),
	).Methods(http.MethodDelete)

	// Third-party auth callbacks
	router.HandleFunc(
		"/v2/session/auth",
		s.authenticate, // No filters applied to this request
	).Methods(http.MethodGet)
}

func (s *SessionsEndpoints) create(w http.ResponseWriter, r *http.Request) {
	// nolint: errcheck
	rootSessionRequest, _ := strconv.ParseBool(r.URL.Query().Get("root"))

	if rootSessionRequest {
		restmachinery.ServeRequest(
			restmachinery.InboundRequest{
				W: w,
				R: r,
				EndpointLogic: func() (interface{}, error) {
					username, password, ok := r.BasicAuth()
					if !ok {
						return nil, &meta.ErrBadRequest{
							Reason: "The request to create a new root session did not " +
								"include a valid basic auth header.",
						}
					}
					return s.Service.CreateRootSession(r.Context(), username, password)
				},
				SuccessCode: http.StatusCreated,
			},
		)
		return
	}

	thirdPartyAuthOpts := &api.ThirdPartyAuthOptions{
		SuccessURL: r.URL.Query().Get("successURL"),
	}
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return s.Service.CreateUserSession(r.Context(), thirdPartyAuthOpts)
			},
			SuccessCode: http.StatusCreated,
		},
	)
}

func (s *SessionsEndpoints) delete(w http.ResponseWriter, r *http.Request) {
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				sessionID := api.SessionIDFromContext(r.Context())
				if sessionID == "" {
					return nil, errors.New(
						"error: delete session request authenticated, but no session ID " +
							"found in request context",
					)
				}
				return nil, s.Service.Delete(r.Context(), sessionID)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (s *SessionsEndpoints) authenticate(
	w http.ResponseWriter,
	r *http.Request,
) {
	defer r.Body.Close() // nolint: errcheck

	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	restmachinery.ServeWebUIRequest(restmachinery.InboundWebUIRequest{
		W: w,
		EndpointLogic: func() (interface{}, error) {
			if state == "" || code == "" {
				return nil, &meta.ErrBadRequest{
					Reason: `The third-party authentication completion request ` +
						`lacked one or both of the "state" and "code" query parameters.`,
				}
			}
			authSuccessURL, err := s.Service.Authenticate(
				r.Context(),
				state,
				code,
			)
			if err != nil {
				return nil,
					errors.Wrap(err, "error completing third-party authentication")
			}
			if authSuccessURL != "" {
				http.Redirect(w, r, authSuccessURL, http.StatusMovedPermanently)
			}
			http.Redirect(w, r, "/assets/splash.html", http.StatusMovedPermanently)
			return nil, nil
		},
		SuccessCode: http.StatusOK,
	})
}
