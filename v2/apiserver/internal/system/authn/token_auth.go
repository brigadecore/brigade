package authn

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authn"
	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	libAuthn "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authn"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/crypto"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

// TokenAuthFilterConfig encapsulates several configuration options for the
// TokenAuthFilter.
type TokenAuthFilterConfig struct {
	// RootUserEnabled indicates whether the TokenAuthFilter should permit the
	// "root" user to authenticate using a password.
	RootUserEnabled bool
	// ThirdPartyAuthEnabled indicates whether the TokenAuthFilter service should
	// permit User authentication via third-parties.
	ThirdPartyAuthEnabled bool
	// FindUserFn is a function for locating a User. This field is applicable only
	// when the value of the ThirdPartyAuthEnabled field is true.
	FindUserFn func(ctx context.Context, id string) (authn.User, error)
	// HashedSchedulerToken is a secure hash of the token used by the scheduler
	// component.
	HashedSchedulerToken string
	// HashedObserverToken is a secure hash of the token used by the observer
	// component.
	HashedObserverToken string
}

// tokenAuthFilter is an implementation of the restmachinery.Filter interface
// that decorates an http.HandlerFunc to carry out request authentication by
// extracting an opaque bearer token form the HTTP Authorization header and
// using that token to locate an established Session.
type tokenAuthFilter struct {
	findServiceAccountByTokenFn func(
		ctx context.Context,
		token string,
	) (authn.ServiceAccount, error)
	findSessionByTokenFn func(
		ctx context.Context,
		token string,
	) (authn.Session, error)
	findEventByTokenFn func(
		ctx context.Context,
		token string,
	) (core.Event, error)
	config TokenAuthFilterConfig
}

// NewTokenAuthFilter returns an implementation of the restmachinery.Filter
// interface that decorates an http.HandlerFunc to carry out request
// authentication by extracting an opaque bearer token form the HTTP
// Authorization header and using that token to locate an established Session.
func NewTokenAuthFilter(
	findServiceAccountByTokenFn func(
		ctx context.Context,
		token string,
	) (authn.ServiceAccount, error),
	findSessionFn func(ctx context.Context, token string) (authn.Session, error),
	findEventByTokenFn func(
		ctx context.Context,
		token string,
	) (core.Event, error),
	config *TokenAuthFilterConfig,
) restmachinery.Filter {
	if config == nil {
		config = &TokenAuthFilterConfig{}
	}
	return &tokenAuthFilter{
		findServiceAccountByTokenFn: findServiceAccountByTokenFn,
		findSessionByTokenFn:        findSessionFn,
		findEventByTokenFn:          findEventByTokenFn,
		config:                      *config,
	}
}

// Decorate decorates one http.HandlerFunc with another.
// nolint: gocyclo
func (t *tokenAuthFilter) Decorate(handle http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		headerValue := r.Header.Get("Authorization")
		if headerValue == "" {
			t.writeResponse(
				w,
				http.StatusUnauthorized,
				&meta.ErrAuthentication{
					Reason: `"Authorization" header is missing.`,
				},
			)
			return
		}
		headerValueParts := strings.SplitN(
			r.Header.Get("Authorization"),
			" ",
			2,
		)
		if len(headerValueParts) != 2 || headerValueParts[0] != "Bearer" {
			t.writeResponse(
				w,
				http.StatusUnauthorized,
				&meta.ErrAuthentication{
					Reason: `"Authorization" header is malformed.`,
				},
			)
			return
		}
		token := headerValueParts[1]

		// Is it the Scheduler's token?
		if crypto.Hash("", token) == t.config.HashedSchedulerToken {
			ctx := libAuthn.ContextWithPrincipal(r.Context(), scheduler)
			handle(w, r.WithContext(ctx))
			return
		}

		// Is it the Observer's token?
		if crypto.Hash("", token) == t.config.HashedObserverToken {
			ctx := libAuthn.ContextWithPrincipal(r.Context(), observer)
			handle(w, r.WithContext(ctx))
			return
		}

		// Is it a Worker's token?
		if event, err := t.findEventByTokenFn(r.Context(), token); err != nil {
			if _, ok := errors.Cause(err).(*meta.ErrNotFound); !ok {
				log.Println(err)
				t.writeResponse(
					w,
					http.StatusInternalServerError,
					&meta.ErrInternalServer{},
				)
				return
			}
		} else {
			ctx := libAuthn.ContextWithPrincipal(r.Context(), worker(event.ID))
			handle(w, r.WithContext(ctx))
			return
		}

		// Is it a ServiceAccount's token?
		if serviceAccount, err :=
			t.findServiceAccountByTokenFn(r.Context(), token); err != nil {
			if _, ok := errors.Cause(err).(*meta.ErrNotFound); !ok {
				log.Println(err)
				t.writeResponse(
					w,
					http.StatusInternalServerError,
					&meta.ErrInternalServer{},
				)
				return
			}
		} else {
			if serviceAccount.Locked != nil {
				http.Error(w, "{}", http.StatusForbidden)
				return
			}
			ctx := libAuthn.ContextWithPrincipal(r.Context(), &serviceAccount)
			handle(w, r.WithContext(ctx))
			return
		}

		session, err := t.findSessionByTokenFn(r.Context(), token)
		if err != nil {
			if _, ok := errors.Cause(err).(*meta.ErrNotFound); ok {
				t.writeResponse(
					w,
					http.StatusUnauthorized,
					&meta.ErrAuthentication{
						Reason: "Session not found or may have expired. " +
							"Please log in again.",
					},
				)
				return
			}
			log.Println(err)
			t.writeResponse(
				w,
				http.StatusInternalServerError,
				&meta.ErrInternalServer{},
			)
			return
		}
		if session.Root && !t.config.RootUserEnabled {
			t.writeResponse(
				w,
				http.StatusUnauthorized,
				&meta.ErrAuthentication{
					Reason: "Supplied token was for an established root session, but " +
						"authentication using root credentials is no longer supported " +
						"by this server.",
				},
			)
			return
		}
		if !session.Root && !t.config.ThirdPartyAuthEnabled {
			t.writeResponse(
				w,
				http.StatusUnauthorized,
				&meta.ErrAuthentication{
					Reason: "Supplied token was for a session whose owner " +
						"authenticated using a third-party, but authentication using " +
						"a third-party is no longer supported by this server.",
				},
			)
			return
		}
		if session.Authenticated == nil {
			t.writeResponse(
				w,
				http.StatusUnauthorized,
				&meta.ErrAuthentication{
					Reason: "Supplied token has not been authenticated. Please log " +
						"in again.",
				},
			)
			return
		}
		// Note: although we have records in the sessions collection configured
		// for TTL cleanup by mongo (per Expires), we wish to keep this check
		// in case this setting is configurable in the future or mongo otherwise
		// fails to clean them up.
		if session.Expires != nil && time.Now().UTC().After(*session.Expires) {
			t.writeResponse(
				w,
				http.StatusUnauthorized,
				&meta.ErrAuthentication{
					Reason: "Supplied token has expired. Please log in again.",
				},
			)
			return
		}
		var principal interface{}
		if session.Root {
			principal = root
		} else {
			user, err := t.config.FindUserFn(r.Context(), session.UserID)
			if err != nil {
				log.Println(err)
				// There should never be an authenticated session for a user that
				// doesn't exist.
				t.writeResponse(
					w,
					http.StatusInternalServerError,
					&meta.ErrInternalServer{},
				)
				return
			}
			if user.Locked != nil {
				http.Error(w, "{}", http.StatusForbidden)
				return
			}
			principal = &user
		}

		// Success! Add the user and the session ID to the context.
		ctx := libAuthn.ContextWithPrincipal(r.Context(), principal)
		ctx = authn.ContextWithSessionID(ctx, session.ID)
		handle(w, r.WithContext(ctx))
	}
}

func (t *tokenAuthFilter) writeResponse(
	w http.ResponseWriter,
	statusCode int,
	response interface{},
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	responseBody, ok := response.([]byte)
	if !ok {
		var err error
		if responseBody, err = json.Marshal(response); err != nil {
			log.Println(errors.Wrap(err, "error marshaling response body"))
		}
	}
	if _, err := w.Write(responseBody); err != nil {
		log.Println(errors.Wrap(err, "error writing response body"))
	}
}
