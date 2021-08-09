package rest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brigadecore/brigade-foundations/crypto"
	"github.com/brigadecore/brigade/v2/apiserver/internal/api"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	const testSessionID = "123456789"

	testCases := []struct {
		name       string
		filter     *tokenAuthFilter
		handler    func(w http.ResponseWriter, r *http.Request)
		setup      func() *http.Request
		assertions func(handlerCalled bool, rr *httptest.ResponseRecorder)
	}{
		{
			name:   "auth header missing",
			filter: &tokenAuthFilter{},
			setup: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				return req
			},
			assertions: func(handlerCalled bool, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rr.Code)
				assert.False(t, handlerCalled)
			},
		},

		{
			name:   "auth header not bearer",
			filter: &tokenAuthFilter{},
			setup: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add("Authorization", "Digest foo")
				return req
			},
			assertions: func(handlerCalled bool, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rr.Code)
				assert.False(t, handlerCalled)
			},
		},

		{
			name: "token belongs to scheduler",
			filter: &tokenAuthFilter{
				config: TokenAuthFilterConfig{
					HashedSchedulerToken: crypto.Hash("", "foo"),
				},
			},
			setup: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add("Authorization", "Bearer foo")
				return req
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				principal := api.PrincipalFromContext(r.Context())
				require.NotNil(t, principal)
				require.Same(t, scheduler, principal)
			},
			assertions: func(handlerCalled bool, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rr.Code)
				assert.True(t, handlerCalled)
			},
		},

		{
			name: "token belongs to observer",
			filter: &tokenAuthFilter{
				config: TokenAuthFilterConfig{
					HashedObserverToken: crypto.Hash("", "foo"),
				},
			},
			setup: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add("Authorization", "Bearer foo")
				return req
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				principal := api.PrincipalFromContext(r.Context())
				require.NotNil(t, principal)
				require.Same(t, observer, principal)
			},
			assertions: func(handlerCalled bool, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rr.Code)
				assert.True(t, handlerCalled)
			},
		},

		{
			name: "error finding event by worker token",
			filter: &tokenAuthFilter{
				findEventByTokenFn: func(context.Context, string) (api.Event, error) {
					return api.Event{}, errors.New("something went wrong")
				},
			},
			setup: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add("Authorization", "Bearer foo")
				return req
			},
			handler: func(w http.ResponseWriter, r *http.Request) {},
			assertions: func(handlerCalled bool, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.False(t, handlerCalled)
			},
		},

		{
			name: "token belongs to a worker",
			filter: &tokenAuthFilter{
				findEventByTokenFn: func(context.Context, string) (api.Event, error) {
					return api.Event{}, nil
				},
			},
			setup: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add("Authorization", "Bearer foo")
				return req
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				principal := api.PrincipalFromContext(r.Context())
				require.NotNil(t, principal)
				require.IsType(t, &workerPrincipal{}, principal)
			},
			assertions: func(handlerCalled bool, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rr.Code)
				assert.True(t, handlerCalled)
			},
		},

		{
			name: "error finding service account",
			filter: &tokenAuthFilter{
				findEventByTokenFn: func(context.Context, string) (api.Event, error) {
					return api.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (api.ServiceAccount, error) {
					return api.ServiceAccount{}, errors.New("something went wrong")
				},
			},
			setup: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add("Authorization", "Bearer foo")
				return req
			},
			assertions: func(handlerCalled bool, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, rr.Code)
				require.False(t, handlerCalled)
			},
		},

		{
			name: "service account found; locked",
			filter: &tokenAuthFilter{
				findEventByTokenFn: func(context.Context, string) (api.Event, error) {
					return api.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (api.ServiceAccount, error) {
					now := time.Now().UTC()
					return api.ServiceAccount{
						Locked: &now,
					}, nil
				},
			},
			setup: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add("Authorization", "Bearer foo")
				return req
			},
			assertions: func(handlerCalled bool, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, rr.Code)
				assert.False(t, handlerCalled)
			},
		},

		{
			name: "service account found; not locked",
			filter: &tokenAuthFilter{
				findEventByTokenFn: func(context.Context, string) (api.Event, error) {
					return api.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (api.ServiceAccount, error) {
					return api.ServiceAccount{}, nil
				},
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				principal := api.PrincipalFromContext(r.Context())
				require.NotNil(t, principal)
				require.IsType(t, &api.ServiceAccount{}, principal)
			},
			setup: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add("Authorization", "Bearer foo")
				return req
			},
			assertions: func(handlerCalled bool, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rr.Code)
				assert.True(t, handlerCalled)
			},
		},

		{
			name: "error finding session",
			filter: &tokenAuthFilter{
				findEventByTokenFn: func(context.Context, string) (api.Event, error) {
					return api.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (api.ServiceAccount, error) {
					return api.ServiceAccount{}, &meta.ErrNotFound{}
				},
				findSessionByTokenFn: func(
					context.Context,
					string,
				) (api.Session, error) {
					return api.Session{}, errors.New("something went wrong")
				},
			},
			setup: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add("Authorization", "Bearer foo")
				return req
			},
			assertions: func(handlerCalled bool, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.False(t, handlerCalled)
			},
		},

		{
			name: "session not found",
			filter: &tokenAuthFilter{
				findEventByTokenFn: func(context.Context, string) (api.Event, error) {
					return api.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (api.ServiceAccount, error) {
					return api.ServiceAccount{}, &meta.ErrNotFound{}
				},
				findSessionByTokenFn: func(
					context.Context,
					string,
				) (api.Session, error) {
					return api.Session{}, &meta.ErrNotFound{}
				},
			},
			setup: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add("Authorization", "Bearer foo")
				return req
			},
			assertions: func(handlerCalled bool, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rr.Code)
				assert.False(t, handlerCalled)
			},
		},

		{
			name: "root session found; root access disabled",
			filter: &tokenAuthFilter{
				findEventByTokenFn: func(context.Context, string) (api.Event, error) {
					return api.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (api.ServiceAccount, error) {
					return api.ServiceAccount{}, &meta.ErrNotFound{}
				},
				findSessionByTokenFn: func(
					context.Context,
					string,
				) (api.Session, error) {
					return api.Session{
						Root: true,
					}, nil
				},
			},
			setup: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add("Authorization", "Bearer foo")
				return req
			},
			assertions: func(handlerCalled bool, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rr.Code)
				assert.False(t, handlerCalled)
			},
		},

		{
			name: "root session found; success",
			filter: &tokenAuthFilter{
				config: TokenAuthFilterConfig{
					RootUserEnabled: true,
				},
				findEventByTokenFn: func(context.Context, string) (api.Event, error) {
					return api.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (api.ServiceAccount, error) {
					return api.ServiceAccount{}, &meta.ErrNotFound{}
				},
				findSessionByTokenFn: func(
					context.Context,
					string,
				) (api.Session, error) {
					auth := time.Now().UTC()
					exp := auth.Add(time.Hour)
					return api.Session{
						Root:          true,
						Authenticated: &auth,
						Expires:       &exp,
					}, nil
				},
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				principal := api.PrincipalFromContext(r.Context())
				require.NotNil(t, principal)
				require.Same(t, root, principal)
			},
			setup: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add("Authorization", "Bearer foo")
				return req
			},
			assertions: func(handlerCalled bool, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rr.Code)
				assert.True(t, handlerCalled)
			},
		},

		{
			name: "user session found; oidc disabled",
			filter: &tokenAuthFilter{
				findEventByTokenFn: func(context.Context, string) (api.Event, error) {
					return api.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (api.ServiceAccount, error) {
					return api.ServiceAccount{}, &meta.ErrNotFound{}
				},
				findSessionByTokenFn: func(
					context.Context,
					string,
				) (api.Session, error) {
					return api.Session{}, nil
				},
			},
			setup: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add("Authorization", "Bearer foo")
				return req
			},
			assertions: func(handlerCalled bool, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rr.Code)
				assert.False(t, handlerCalled)
			},
		},

		{
			name: "user session found; token not activated",
			filter: &tokenAuthFilter{
				config: TokenAuthFilterConfig{
					ThirdPartyAuthEnabled: true,
				},
				findEventByTokenFn: func(context.Context, string) (api.Event, error) {
					return api.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (api.ServiceAccount, error) {
					return api.ServiceAccount{}, &meta.ErrNotFound{}
				},
				findSessionByTokenFn: func(
					context.Context,
					string,
				) (api.Session, error) {
					return api.Session{}, nil
				},
			},
			setup: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add("Authorization", "Bearer foo")
				return req
			},
			assertions: func(handlerCalled bool, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rr.Code)
				assert.False(t, handlerCalled)
			},
		},

		{
			name: "user session found; token expired",
			filter: &tokenAuthFilter{
				config: TokenAuthFilterConfig{
					ThirdPartyAuthEnabled: true,
				},
				findEventByTokenFn: func(context.Context, string) (api.Event, error) {
					return api.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (api.ServiceAccount, error) {
					return api.ServiceAccount{}, &meta.ErrNotFound{}
				},
				findSessionByTokenFn: func(
					context.Context,
					string,
				) (api.Session, error) {
					auth := time.Now().UTC().Add(-2 * time.Hour)
					exp := auth.Add(time.Hour)
					return api.Session{
						Authenticated: &auth,
						Expires:       &exp,
					}, nil
				},
			},
			setup: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add("Authorization", "Bearer foo")
				return req
			},
			assertions: func(handlerCalled bool, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rr.Code)
				assert.False(t, handlerCalled)
			},
		},

		{
			name: "user session found; token valid; error finding user",
			filter: &tokenAuthFilter{
				config: TokenAuthFilterConfig{
					ThirdPartyAuthEnabled: true,
					FindUserFn: func(ctx context.Context, id string) (api.User, error) {
						return api.User{}, errors.New("something went wrong")
					},
				},
				findEventByTokenFn: func(context.Context, string) (api.Event, error) {
					return api.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (api.ServiceAccount, error) {
					return api.ServiceAccount{}, &meta.ErrNotFound{}
				},
				findSessionByTokenFn: func(
					context.Context,
					string,
				) (api.Session, error) {
					auth := time.Now().UTC()
					exp := auth.Add(time.Hour)
					return api.Session{
						Authenticated: &auth,
						Expires:       &exp,
					}, nil
				},
			},
			setup: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add("Authorization", "Bearer foo")
				return req
			},
			assertions: func(handlerCalled bool, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.False(t, handlerCalled)
			},
		},

		{
			name: "user session found; token valid; user locked",
			filter: &tokenAuthFilter{
				config: TokenAuthFilterConfig{
					ThirdPartyAuthEnabled: true,
					FindUserFn: func(ctx context.Context, id string) (api.User, error) {
						now := time.Now().UTC()
						return api.User{
							Locked: &now,
						}, nil
					},
				},
				findEventByTokenFn: func(context.Context, string) (api.Event, error) {
					return api.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (api.ServiceAccount, error) {
					return api.ServiceAccount{}, &meta.ErrNotFound{}
				},
				findSessionByTokenFn: func(
					context.Context,
					string,
				) (api.Session, error) {
					auth := time.Now().UTC()
					exp := auth.Add(time.Hour)
					return api.Session{
						Authenticated: &auth,
						Expires:       &exp,
					}, nil
				},
			},
			setup: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add("Authorization", "Bearer foo")
				return req
			},
			assertions: func(handlerCalled bool, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, rr.Code)
				assert.False(t, handlerCalled)
			},
		},

		{
			name: "user session found; success",
			filter: &tokenAuthFilter{
				config: TokenAuthFilterConfig{
					ThirdPartyAuthEnabled: true,
					FindUserFn: func(ctx context.Context, id string) (api.User, error) {
						return api.User{}, nil
					},
				},
				findEventByTokenFn: func(context.Context, string) (api.Event, error) {
					return api.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (api.ServiceAccount, error) {
					return api.ServiceAccount{}, &meta.ErrNotFound{}
				},
				findSessionByTokenFn: func(
					context.Context,
					string,
				) (api.Session, error) {
					auth := time.Now().UTC()
					exp := auth.Add(time.Hour)
					return api.Session{
						ObjectMeta: meta.ObjectMeta{
							ID: testSessionID,
						},
						Authenticated: &auth,
						Expires:       &exp,
					}, nil
				},
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				principal := api.PrincipalFromContext(r.Context())
				require.NotNil(t, principal)
				require.IsType(t, &api.User{}, principal)
				sessionID := api.SessionIDFromContext(r.Context())
				require.Equal(t, testSessionID, sessionID)
			},
			setup: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add("Authorization", "Bearer foo")
				return req
			},
			assertions: func(handlerCalled bool, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rr.Code)
				assert.True(t, handlerCalled)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := testCase.setup()
			handlerCalled := false
			testCase.filter.Decorate(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				if testCase.handler != nil {
					testCase.handler(w, r)
				}
			})(rr, req)
			testCase.assertions(handlerCalled, rr)
		})
	}
}
