package authn

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/require"
)

func TestTokenAuthFilterWithHeaderMissing(t *testing.T) {
	filter := &tokenAuthFilter{
		config: TokenAuthFilterConfig{
			OpenIDConnectEnabled: true,
		},
	}
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	rr := httptest.NewRecorder()
	handlerCalled := false
	filter.Decorate(func(http.ResponseWriter, *http.Request) {
		handlerCalled = true
	})(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)
	require.False(t, handlerCalled)
}

func TestTokenAuthFilterWithHeaderNotBearer(t *testing.T) {
	filter := &tokenAuthFilter{
		config: TokenAuthFilterConfig{
			OpenIDConnectEnabled: true,
		},
	}
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	req.Header.Add("Authorization", "Digest foo")
	rr := httptest.NewRecorder()
	handlerCalled := false
	filter.Decorate(func(http.ResponseWriter, *http.Request) {
		handlerCalled = true
	})(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)
	require.False(t, handlerCalled)
}

func TestTokenAuthFilterWithTokenInvalid(t *testing.T) {
	filter := &tokenAuthFilter{
		findSessionFn: func(context.Context, string) (authx.Session, error) {
			return authx.Session{}, &meta.ErrNotFound{}
		},
		config: TokenAuthFilterConfig{
			OpenIDConnectEnabled: true,
		},
	}
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	req.Header.Add(
		"Authorization",
		fmt.Sprintf("Bearer %s", "foo"),
	)
	rr := httptest.NewRecorder()
	handlerCalled := false
	filter.Decorate(func(http.ResponseWriter, *http.Request) {
		handlerCalled = true
	})(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)
	require.False(t, handlerCalled)
}

func TestTokenAuthFilterWithUnauthenticatedSession(t *testing.T) {
	filter := &tokenAuthFilter{
		findSessionFn: func(context.Context, string) (authx.Session, error) {
			return authx.Session{}, nil
		},
		config: TokenAuthFilterConfig{
			OpenIDConnectEnabled: true,
			FindUserFn: func(context.Context, string) (authx.User, error) {
				return authx.User{}, nil
			},
		},
	}
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	req.Header.Add("Authorization", "Bearer foobar")
	rr := httptest.NewRecorder()
	var handlerCalled bool
	filter.Decorate(func(_ http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		require.Nil(t, authx.PrincipalFromContext(r.Context()))
		require.Empty(t, authx.SessionIDFromContext(r.Context()))
	})(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)
	require.False(t, handlerCalled)
}

func TestTokenAuthFilterWithAuthenticatedSession(t *testing.T) {
	const testSessionID = "foobar"
	filter := &tokenAuthFilter{
		findSessionFn: func(context.Context, string) (authx.Session, error) {
			now := time.Now()
			expiry := now.Add(time.Minute)
			return authx.Session{
				ObjectMeta: meta.ObjectMeta{
					ID: testSessionID,
				},
				Authenticated: &now,
				Expires:       &expiry,
			}, nil
		},
		config: TokenAuthFilterConfig{
			OpenIDConnectEnabled: true,
			FindUserFn: func(context.Context, string) (authx.User, error) {
				return authx.User{}, nil
			},
		},
	}
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	req.Header.Add("Authorization", "Bearer foobar")
	rr := httptest.NewRecorder()
	var handlerCalled bool
	filter.Decorate(func(_ http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		require.NotNil(t, authx.PrincipalFromContext(r.Context()))
		require.Equal(t, testSessionID, authx.SessionIDFromContext(r.Context()))
	})(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	require.True(t, handlerCalled)
}
