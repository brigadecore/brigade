package authn

import (
	"context"
	"encoding/json"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/crypto"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/coreos/go-oidc"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/oauth2"
)

// OIDCAuthDetails encapsulates all information required for a client
// authenticating by means of OpenID Connect to complete the authentication
// process using a third-party identity provider.
type OIDCAuthDetails struct {
	// AuthURL is a URL that can be requested in a user's web browser to complete
	// authentication via a third-party OIDC identity provider.
	AuthURL string `json:"authURL"`
	// Token is an opaque bearer token issued by Brigade to correlate a User with
	// a Session. It remains unactivated (useless) until the OIDC authentication
	// workflow is successfully completed. Clients may expect that that the token
	// expires (at an interval determined by a system administrator) and, for
	// simplicity, is NOT refreshable. When the token has expired,
	// re-authentication is required.
	Token string `json:"token"`
}

// MarshalJSON amends OIDCAuthDetails instances with type metadata.
func (o OIDCAuthDetails) MarshalJSON() ([]byte, error) {
	type Alias OIDCAuthDetails
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "OIDCAuthDetails",
			},
			Alias: (Alias)(o),
		},
	)
}

// Session encapsulates details of a session belonging either to the root user
// or a discrete User that has authenticated (or is in the process of
// authenticating) via OpenID Connect.
type Session struct {
	// ObjectMeta encapsulates Session metadata.
	meta.ObjectMeta `json:"metadata" bson:",inline"`
	// Root indicates whether the Session belongs to the root user (true) or a
	// some discrete User.
	Root bool `json:"root" bson:"root"`
	// UserID, if set, identifies the discrete User to whom this Session belongs.
	UserID string `json:"userID" bson:"userID"`
	// HashedOAuth2State, if set, is a secure hash of the OAuth 2 "state" code
	// used in completing authentication via OpenID Connect.
	HashedOAuth2State string `json:"-" bson:"hashedOAuth2State"`
	// HashedToken is a secure hash of the opaque bearer token associated with
	// this Session.
	HashedToken string `json:"-" bson:"hashedToken"`
	// Authenticated indicates the date/time at which authentication was completed
	// successfully. If the value of this field is nil, the Session is NOT
	// authenticated.
	Authenticated *time.Time `json:"authenticated" bson:"authenticated"`
	// Expires, if set, specified an expiry date/time for the Session and its
	// associated token.
	Expires *time.Time `json:"expires" bson:"expires"`
}

type sessionIDContextKey struct{}

// ContextWithSessionID returns a context.Context that has been augmented with
// the provided Session identifier.
func ContextWithSessionID(
	ctx context.Context,
	sessionID string,
) context.Context {
	return context.WithValue(ctx, sessionIDContextKey{}, sessionID)
}

// SessionIDFromContext extracts a Session identifier from the provided
// context.Context and returns it.
func SessionIDFromContext(ctx context.Context) string {
	token := ctx.Value(sessionIDContextKey{})
	if token == nil {
		return ""
	}
	return token.(string)
}

// OAuth2Helper is an interface for the subset of *oauth2.Config functions used
// for Brigade Session management. Dependence on this interface instead of
// directly upon the *oauth2.Config allows for the possibility of utilizing a
// mock implementation for testing purposes. Adding only the subset of functions
// that we actually use limits the effort involved in creating such mocks.
type OAuth2Helper interface {
	// AuthCodeURL given an OAuth 2 state code and oauth2.AuthCodeOption returns
	// the URL that a user may visit with their web browser in order to complete
	// authentication using OpenID Connect.
	AuthCodeURL(
		state string,
		opts ...oauth2.AuthCodeOption,
	) string
	// Exchange exchanges the given OAuth 2 code for an *oauth2.Token.
	Exchange(
		ctx context.Context,
		code string,
		opts ...oauth2.AuthCodeOption,
	) (*oauth2.Token, error)
}

// OpenIDConnectTokenVerifier is an interface for the subset of
// *oidc.IDTokenVerifier used for Brigade Session management. Dependence on this
// interface instead of directly upon the *oidc.IDTokenVerifier allows for the
// possibility of utilizing a mock implementation for testing purposes. Adding
// only the subset of functions that we actually use limits the effort involved
// in creating such mocks.
type OpenIDConnectTokenVerifier interface {
	Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error)
}

// SessionsServiceConfig encapsulates several configuration options for the
// Sessions service.
type SessionsServiceConfig struct {
	// RootUserEnabled indicates whether the Session service should permit the
	// "root" user to authenticate using a password.
	RootUserEnabled bool
	// RootUserPassword specifies the password that must be supplied by users
	// attempting to authenticate as the "root" user. This field is applicable
	// only when value of the RootUserEnabled field is true.
	RootUserPassword string
	// OpenIDConnectEnabled indicates whether the Session service should permit
	// User authentication via OpenID Connect.
	OpenIDConnectEnabled bool
	// OAuth2Helper provides authentication-related functions configured for a
	// specific OpenID Connect identity provider. This field is applicable only
	// when value of the OpenIDConnectEnabled field is true.
	OAuth2Helper OAuth2Helper
	// OpenIDConnectTokenVerifier provides an OpenID Connect token verifier. This
	// field is applicable only when value of the OpenIDConnectEnabled field is
	// true.
	OpenIDConnectTokenVerifier OpenIDConnectTokenVerifier
}

// SessionsService is the specialized interface for managing Sessions. It's
// decoupled from underlying technology choices (e.g. data store) to keep
// business logic reusable and consistent while the underlying tech stack
// remains free to change.
type SessionsService interface {
	// CreateRootSession creates a Session for the root user (if enabled by the
	// system administrator) and returns a Token with a short expiry period
	// (determined by a system administrator). If authentication as the root user
	// is not enabled, implementations MUST return a *meta.ErrNotSupported error.
	// If the specified username is not "root" or the specified password is
	// incorrect, implementations MUST return a *meta.ErrAuthentication error.
	CreateRootSession(
		ctx context.Context,
		username string,
		password string,
	) (Token, error)
	// CreateUserSession creates a new User Session and initiates an OpenID
	// Connect authentication workflow (if authentication using OpenID connect has
	// been enabled by the system administrator). It returns an OIDCAuthDetails
	// containing all information required to continue the authentication process
	// with a third-party OIDC identity provider. If authentication using OpenID
	// Connect is not enabled, implementations MUST return a *meta.ErrNotSupported
	// error.
	CreateUserSession(context.Context) (OIDCAuthDetails, error)
	// Authenticate completes the final steps of the OpenID Connect authentication
	// workflow (if authentication using OpenID connect has been enabled by the
	// system administrator). It uses the provided oauth2State to identify an
	// as-yet anonymous Session (with an as-yet unactivated token). It
	// communicates with the third-party OIDC identity provider, exchanging the
	// provided oidcCode for user information. This information can be used to
	// correlate the as-yet anonymous Session to an existing User. If the User is
	// previously unknown to Brigade, implementations MUST seamlessly create one
	// (with read-only permissions) based on information provided by the identity
	// provider. Finally, the Session's token is activated. If authentication
	// using OpenID Connect is not enabled, implementations MUST return a
	// *meta.ErrNotSupported error.
	Authenticate(
		ctx context.Context,
		oauth2State string,
		oidcCode string,
	) error
	// GetByToken retrieves the Session having the provided token. If no such
	// Session is found or is found but is expired, implementations MUST return a
	// *meta.ErrAuthentication error.
	GetByToken(ctx context.Context, token string) (Session, error)
	// Delete deletes the specified Session.
	Delete(ctx context.Context, id string) error
}

// sessionsService is an implementation of the SessionsService interface.
type sessionsService struct {
	sessionsStore          SessionsStore
	usersStore             UsersStore
	config                 SessionsServiceConfig
	hashedRootUserPassword string
}

// NewSessionsService returns a specialized interface for managing Sessions.
func NewSessionsService(
	sessionsStore SessionsStore,
	usersStore UsersStore,
	config *SessionsServiceConfig,
) SessionsService {
	if config == nil {
		config = &SessionsServiceConfig{}
	}
	svc := &sessionsService{
		sessionsStore: sessionsStore,
		usersStore:    usersStore,
		config:        *config,
	}
	if config.RootUserPassword != "" {
		svc.hashedRootUserPassword = crypto.Hash("root", config.RootUserPassword)
		// Don't let the cleartext password float around in memory longer than
		// needed
		svc.config.RootUserPassword = ""
	}
	return svc
}

func (s *sessionsService) CreateRootSession(
	ctx context.Context,
	username string,
	password string,
) (Token, error) {
	token := Token{
		Value: crypto.NewToken(256),
	}
	if !s.config.RootUserEnabled {
		return token, &meta.ErrNotSupported{
			Details: "Authentication using root credentials is not supported by " +
				"this server.",
		}
	}
	if username != "root" ||
		crypto.Hash(username, password) != s.hashedRootUserPassword {
		return token, &meta.ErrAuthentication{
			Reason: "Could not authenticate request using the supplied credentials.",
		}
	}

	now := time.Now().UTC()
	expiryTime := now.Add(time.Hour)
	session := Session{
		ObjectMeta: meta.ObjectMeta{
			ID:      uuid.NewV4().String(),
			Created: &now,
		},
		Root:          true,
		HashedToken:   crypto.Hash("", token.Value),
		Authenticated: &now,
		Expires:       &expiryTime,
	}

	if err := s.sessionsStore.Create(ctx, session); err != nil {
		return token, errors.Wrapf(
			err,
			"error storing new root session %q",
			session.ID,
		)
	}
	return token, nil
}

func (s *sessionsService) CreateUserSession(
	ctx context.Context,
) (OIDCAuthDetails, error) {
	if !s.config.OpenIDConnectEnabled {
		return OIDCAuthDetails{}, &meta.ErrNotSupported{
			Details: "Authentication using OpenID Connect is not supported by this " +
				"server.",
		}
	}
	oauth2State := crypto.NewToken(30)
	token := crypto.NewToken(256)
	session := Session{
		ObjectMeta: meta.ObjectMeta{
			ID: uuid.NewV4().String(),
		},
		HashedOAuth2State: crypto.Hash("", oauth2State),
		HashedToken:       crypto.Hash("", token),
	}
	now := time.Now().UTC()
	session.Created = &now
	if err := s.sessionsStore.Create(ctx, session); err != nil {
		return OIDCAuthDetails{}, errors.Wrapf(
			err,
			"error storing new user session %q",
			session.ID,
		)
	}
	return OIDCAuthDetails{
		Token:   token,
		AuthURL: s.config.OAuth2Helper.AuthCodeURL(oauth2State),
	}, nil
}

func (s *sessionsService) Authenticate(
	ctx context.Context,
	oauth2State string,
	oidcCode string,
) error {
	if !s.config.OpenIDConnectEnabled {
		return &meta.ErrNotSupported{
			Details: "Authentication using OpenID Connect is not supported by this " +
				"server.",
		}
	}
	session, err := s.sessionsStore.GetByHashedOAuth2State(
		ctx,
		crypto.Hash("", oauth2State),
	)
	if err != nil {
		return errors.Wrap(
			err,
			"error retrieving session from store by hashed OAuth2 state",
		)
	}
	oauth2Token, err := s.config.OAuth2Helper.Exchange(ctx, oidcCode)
	if err != nil {
		return errors.Wrap(
			err,
			"error exchanging OpenID Connect code for OAuth2 token",
		)
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return errors.New(
			"OAuth2 token, did not include an OpenID Connect identity token",
		)
	}
	idToken, err := s.config.OpenIDConnectTokenVerifier.Verify(ctx, rawIDToken)
	if err != nil {
		return errors.Wrap(err, "error verifying OpenID Connect identity token")
	}
	claims := struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}{}
	if err = idToken.Claims(&claims); err != nil {
		return errors.Wrap(
			err,
			"error decoding OpenID Connect identity token claims",
		)
	}
	user, err := s.usersStore.Get(ctx, claims.Email)
	if err != nil {
		if _, ok := errors.Cause(err).(*meta.ErrNotFound); ok {
			// User wasn't found. That's ok. We'll create one.
			user = User{
				ObjectMeta: meta.ObjectMeta{
					ID: claims.Email,
				},
				Name: claims.Name,
			}
			if err = s.usersStore.Create(ctx, user); err != nil {
				return errors.Wrapf(err, "error storing new user %q", user.ID)
			}
		} else {
			// It was something else that went wrong when searching for the user.
			return err
		}
	}
	if err := s.sessionsStore.Authenticate(
		ctx,
		session.ID,
		user.ID,
		time.Now().UTC().Add(time.Hour),
	); err != nil {
		return errors.Wrapf(
			err,
			"error storing authentication details for session %q",
			session.ID,
		)
	}
	return nil
}

func (s *sessionsService) GetByToken(
	ctx context.Context,
	token string,
) (Session, error) {
	session, err := s.sessionsStore.GetByHashedToken(
		ctx,
		crypto.Hash("", token),
	)
	if err != nil {
		return session, errors.Wrap(
			err,
			"error retrieving session from store by hashed token",
		)
	}
	return session, nil
}

func (s *sessionsService) Delete(ctx context.Context, id string) error {
	if err := s.sessionsStore.Delete(ctx, id); err != nil {
		return errors.Wrapf(err, "error removing session %q from store", id)
	}
	return nil
}

// SessionsStore is an interface for Session persistence operations.
type SessionsStore interface {
	// Create stores the provided Session. Implementations MUST return an error if
	// a Session having the indicated identifier already exists.
	Create(context.Context, Session) error
	// GetByHashedOAuth2State returns a Session having the indicated secure hash
	// of the OAuth 2 "state" code. This is used in completing the OpenID Connect
	// authentication workflow. If no such Session exists, implementations MUST
	// return a *meta.ErrNotFound error.
	GetByHashedOAuth2State(context.Context, string) (Session, error)
	// GetByHashedToken returns a Session having the indicated secure hash of the
	// opaque bearer token. If no such Session exists, implementations MUST
	// return a *meta.ErrNotFound error.
	GetByHashedToken(context.Context, string) (Session, error)
	// Authenticate updates the specified, as-yet-anonymous Session (with an
	// as-yet unactivated token) to denote ownership by the indicated User and to
	// assign the specified expiry date/time. This is used in completing the
	// OpenID Connect authentication workflow.
	Authenticate(
		ctx context.Context,
		sessionID string,
		userID string,
		expires time.Time,
	) error
	// Delete deletes the specified Session. If no Session having the given
	// identifier is found, implementations MUST return a *meta.ErrNotFound error.
	Delete(context.Context, string) error
	// DeleteByUser deletes all sessions belonging to the specified User.
	DeleteByUser(ctx context.Context, userID string) error
}
