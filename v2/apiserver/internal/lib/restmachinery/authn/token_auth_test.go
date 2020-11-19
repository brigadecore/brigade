package authn

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/crypto"
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
				principal := authx.PrincipalFromContext(r.Context())
				require.NotNil(t, principal)
				require.Same(t, authx.Scheduler, principal)
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
				principal := authx.PrincipalFromContext(r.Context())
				require.NotNil(t, principal)
				require.Same(t, authx.Observer, principal)
			},
			assertions: func(handlerCalled bool, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rr.Code)
				assert.True(t, handlerCalled)
			},
		},

		{
			name: "error finding event by worker token",
			filter: &tokenAuthFilter{
				findEventByTokenFn: func(context.Context, string) (core.Event, error) {
					return core.Event{}, errors.New("something went wrong")
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
				findEventByTokenFn: func(context.Context, string) (core.Event, error) {
					return core.Event{}, nil
				},
			},
			setup: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add("Authorization", "Bearer foo")
				return req
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				principal := authx.PrincipalFromContext(r.Context())
				require.NotNil(t, principal)
				require.Equal(t, "*authx.worker", reflect.TypeOf(principal).String())
			},
			assertions: func(handlerCalled bool, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rr.Code)
				assert.True(t, handlerCalled)
			},
		},

		{
			name: "error finding service account",
			filter: &tokenAuthFilter{
				findEventByTokenFn: func(context.Context, string) (core.Event, error) {
					return core.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (authx.ServiceAccount, error) {
					return authx.ServiceAccount{}, errors.New("something went wrong")
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
				findEventByTokenFn: func(context.Context, string) (core.Event, error) {
					return core.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (authx.ServiceAccount, error) {
					now := time.Now().UTC()
					return authx.ServiceAccount{
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
				findEventByTokenFn: func(context.Context, string) (core.Event, error) {
					return core.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (authx.ServiceAccount, error) {
					return authx.ServiceAccount{}, nil
				},
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				principal := authx.PrincipalFromContext(r.Context())
				require.NotNil(t, principal)
				require.IsType(t, &authx.ServiceAccount{}, principal)
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
				findEventByTokenFn: func(context.Context, string) (core.Event, error) {
					return core.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (authx.ServiceAccount, error) {
					return authx.ServiceAccount{}, &meta.ErrNotFound{}
				},
				findSessionByTokenFn: func(
					context.Context,
					string,
				) (authx.Session, error) {
					return authx.Session{}, errors.New("something went wrong")
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
				findEventByTokenFn: func(context.Context, string) (core.Event, error) {
					return core.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (authx.ServiceAccount, error) {
					return authx.ServiceAccount{}, &meta.ErrNotFound{}
				},
				findSessionByTokenFn: func(
					context.Context,
					string,
				) (authx.Session, error) {
					return authx.Session{}, &meta.ErrNotFound{}
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
				findEventByTokenFn: func(context.Context, string) (core.Event, error) {
					return core.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (authx.ServiceAccount, error) {
					return authx.ServiceAccount{}, &meta.ErrNotFound{}
				},
				findSessionByTokenFn: func(
					context.Context,
					string,
				) (authx.Session, error) {
					return authx.Session{
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
				findEventByTokenFn: func(context.Context, string) (core.Event, error) {
					return core.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (authx.ServiceAccount, error) {
					return authx.ServiceAccount{}, &meta.ErrNotFound{}
				},
				findSessionByTokenFn: func(
					context.Context,
					string,
				) (authx.Session, error) {
					auth := time.Now().UTC()
					exp := auth.Add(time.Hour)
					return authx.Session{
						Root:          true,
						Authenticated: &auth,
						Expires:       &exp,
					}, nil
				},
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				principal := authx.PrincipalFromContext(r.Context())
				require.NotNil(t, principal)
				require.Same(t, authx.Root, principal)
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
				findEventByTokenFn: func(context.Context, string) (core.Event, error) {
					return core.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (authx.ServiceAccount, error) {
					return authx.ServiceAccount{}, &meta.ErrNotFound{}
				},
				findSessionByTokenFn: func(
					context.Context,
					string,
				) (authx.Session, error) {
					return authx.Session{}, nil
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
					OpenIDConnectEnabled: true,
				},
				findEventByTokenFn: func(context.Context, string) (core.Event, error) {
					return core.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (authx.ServiceAccount, error) {
					return authx.ServiceAccount{}, &meta.ErrNotFound{}
				},
				findSessionByTokenFn: func(
					context.Context,
					string,
				) (authx.Session, error) {
					return authx.Session{}, nil
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
					OpenIDConnectEnabled: true,
				},
				findEventByTokenFn: func(context.Context, string) (core.Event, error) {
					return core.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (authx.ServiceAccount, error) {
					return authx.ServiceAccount{}, &meta.ErrNotFound{}
				},
				findSessionByTokenFn: func(
					context.Context,
					string,
				) (authx.Session, error) {
					auth := time.Now().UTC().Add(-2 * time.Hour)
					exp := auth.Add(time.Hour)
					return authx.Session{
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
					OpenIDConnectEnabled: true,
					FindUserFn: func(ctx context.Context, id string) (authx.User, error) {
						return authx.User{}, errors.New("something went wrong")
					},
				},
				findEventByTokenFn: func(context.Context, string) (core.Event, error) {
					return core.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (authx.ServiceAccount, error) {
					return authx.ServiceAccount{}, &meta.ErrNotFound{}
				},
				findSessionByTokenFn: func(
					context.Context,
					string,
				) (authx.Session, error) {
					auth := time.Now().UTC()
					exp := auth.Add(time.Hour)
					return authx.Session{
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
					OpenIDConnectEnabled: true,
					FindUserFn: func(ctx context.Context, id string) (authx.User, error) {
						now := time.Now().UTC()
						return authx.User{
							Locked: &now,
						}, nil
					},
				},
				findEventByTokenFn: func(context.Context, string) (core.Event, error) {
					return core.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (authx.ServiceAccount, error) {
					return authx.ServiceAccount{}, &meta.ErrNotFound{}
				},
				findSessionByTokenFn: func(
					context.Context,
					string,
				) (authx.Session, error) {
					auth := time.Now().UTC()
					exp := auth.Add(time.Hour)
					return authx.Session{
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
					OpenIDConnectEnabled: true,
					FindUserFn: func(ctx context.Context, id string) (authx.User, error) {
						return authx.User{}, nil
					},
				},
				findEventByTokenFn: func(context.Context, string) (core.Event, error) {
					return core.Event{}, &meta.ErrNotFound{}
				},
				findServiceAccountByTokenFn: func(
					context.Context,
					string,
				) (authx.ServiceAccount, error) {
					return authx.ServiceAccount{}, &meta.ErrNotFound{}
				},
				findSessionByTokenFn: func(
					context.Context,
					string,
				) (authx.Session, error) {
					auth := time.Now().UTC()
					exp := auth.Add(time.Hour)
					return authx.Session{
						ObjectMeta: meta.ObjectMeta{
							ID: testSessionID,
						},
						Authenticated: &auth,
						Expires:       &exp,
					}, nil
				},
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				principal := authx.PrincipalFromContext(r.Context())
				require.NotNil(t, principal)
				require.IsType(t, &authx.User{}, principal)
				sessionID := authx.SessionIDFromContext(r.Context())
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
