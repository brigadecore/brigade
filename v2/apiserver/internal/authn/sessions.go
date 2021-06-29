package authn

import (
	"context"
	"encoding/json"
	"time"

	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/crypto"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/brigadecore/brigade/v2/apiserver/internal/system"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type ThirdPartyAuthStrategy string

// SessionKind represents the canonical Session kind string
const SessionKind = "Session"

// ThirdPartyAuthOptions encapsulates user-specified options when creating a new
// Session that will authenticate using a third-party identity provider.
type ThirdPartyAuthOptions struct {
	// SuccessURL indicates where users should be redirected to after successful
	// completion of a third-party authentication workflow. If this is left
	// unspecified, users will be redirected to a default success page.
	SuccessURL string
}

// ThirdPartyAuthDetails encapsulates all information required for a client
// authenticating by means of a third-party identity provider to complete the
// authentication workflow.
type ThirdPartyAuthDetails struct {
	// AuthURL is a URL that can be requested in a user's web browser to complete
	// authentication via a third-party identity provider.
	AuthURL string `json:"authURL"`
	// Token is an opaque bearer token issued by Brigade to correlate a User with
	// a Session. It remains unactivated (useless) until the authentication
	// workflow is successfully completed. Clients may expect that that the token
	// expires (at an interval determined by a system administrator) and, for
	// simplicity, is NOT refreshable. When the token has expired,
	// re-authentication is required.
	Token string `json:"token"`
}

// MarshalJSON amends ThirdPartyAuthDetails instances with type metadata.
func (t ThirdPartyAuthDetails) MarshalJSON() ([]byte, error) {
	type Alias ThirdPartyAuthDetails
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "ThirdPartyAuthDetails",
			},
			Alias: (Alias)(t),
		},
	)
}

// Session encapsulates details of a session belonging either to the root user
// or a discrete User that has authenticated (or is in the process of
// authenticating) via OpenID Connect or GitHub.
type Session struct {
	// ObjectMeta encapsulates Session metadata.
	meta.ObjectMeta `json:"metadata" bson:",inline"`
	// Root indicates whether the Session belongs to the root user (true) or a
	// some discrete User.
	Root bool `json:"root" bson:"root"`
	// UserID, if set, identifies the discrete User to whom this Session belongs.
	UserID string `json:"userID" bson:"userID"`
	// HashedOAuth2State, if set, is a secure hash of the OAuth 2 "state" code
	// used in completing authentication via a third-party identity provider.
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
	// AuthSuccessURL indicates a URL to redirect the User to after successful
	// completion of a third-party authentication workflow. If not specified, a
	// default URL is used.
	AuthSuccessURL string `json:"authSuccessURL" bson:"authSuccessURL"`
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

// SessionsServiceConfig encapsulates several configuration options for the
// Sessions service.
type SessionsServiceConfig struct {
	// RootUserEnabled indicates whether the Session service should permit the
	// "root" user to authenticate using a password.
	RootUserEnabled bool
	// RootUserSessionTTL specifies the TTL for the root user session. This value
	// will be used to set the Expires field on the Session record for the root
	// user.
	RootUserSessionTTL time.Duration
	// RootUserPassword specifies the password that must be supplied by users
	// attempting to authenticate as the "root" user. This field is applicable
	// only when value of the RootUserEnabled field is true.
	RootUserPassword string
	// ThirdPartyAuthEnabled indicates whether authentication using a third-party
	// identity provider is supported by the Sessions service.
	ThirdPartyAuthEnabled bool
	// UserSessionTTL specifies the TTL for user sessions. This value will be
	// used to set the Expires field on the Session record for each user.
	UserSessionTTL time.Duration
	// AdminUserIDs enumerates users who should be granted system admin privileges
	// the first time they log in.
	AdminUserIDs []string
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
	// CreateUserSession creates a new User Session and initiates a third-party
	// authentication workflow (if enabled by the system administrator). It
	// returns ThirdPartyAuthDetails containing all information required to
	// continue the authentication process with the third-party identity provider.
	// If authentication using a third-party is not enabled, implementations MUST
	// return a *meta.ErrNotSupported error.
	CreateUserSession(
		context.Context,
		*ThirdPartyAuthOptions,
	) (ThirdPartyAuthDetails, error)
	// Authenticate completes the final steps of the third-party authentication
	// workflow (if enabled by the system administrator) and returns a URL to
	// which the user may be redirected. It uses the provided state to identify an
	// as-yet anonymous Session (with an as-yet unactivated token). It
	// communicates with the third-party identity provider, exchanging the
	// provided code for user information. This information can be used to
	// correlate the as-yet anonymous Session to an existing User. If the User is
	// previously unknown to Brigade, implementations MUST seamlessly create one
	// (with no initial permissions) based on information provided by the identity
	// provider. Finally, the Session's token is activated. If authentication
	// using a third-party is not enabled, implementations MUST return a
	// *meta.ErrNotSupported error.
	Authenticate(ctx context.Context, state string, code string) (string, error)
	// GetByToken retrieves the Session having the provided token. If no such
	// Session is found or is found but is expired, implementations MUST return a
	// *meta.ErrAuthentication error.
	GetByToken(ctx context.Context, token string) (Session, error)
	// Delete deletes the specified Session.
	Delete(ctx context.Context, id string) error
}

// sessionsService is an implementation of the SessionsService interface.
type sessionsService struct {
	sessionsStore SessionsStore
	usersStore    UsersStore
	// Instead of getting the whole RoleAssignmentsStore, we get just the one
	// function we need from that store. This is a workaround to avoid an import
	// cycle.
	grantRoleFn            func(context.Context, libAuthz.RoleAssignment) error
	thirdPartyAuthHelper   ThirdPartyAuthHelper
	config                 SessionsServiceConfig
	hashedRootUserPassword string
}

// NewSessionsService returns a specialized interface for managing Sessions.
func NewSessionsService(
	sessionsStore SessionsStore,
	usersStore UsersStore,
	grantRoleFn func(context.Context, libAuthz.RoleAssignment) error,
	thirdPartyAuthHelper ThirdPartyAuthHelper,
	config *SessionsServiceConfig,
) SessionsService {
	if config == nil {
		config = &SessionsServiceConfig{}
	}
	svc := &sessionsService{
		sessionsStore:        sessionsStore,
		usersStore:           usersStore,
		grantRoleFn:          grantRoleFn,
		thirdPartyAuthHelper: thirdPartyAuthHelper,
		config:               *config,
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
	expiryTime := now.Add(s.config.RootUserSessionTTL)
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
	opts *ThirdPartyAuthOptions,
) (ThirdPartyAuthDetails, error) {
	if !s.config.ThirdPartyAuthEnabled {
		return ThirdPartyAuthDetails{}, &meta.ErrNotSupported{
			Details: "Authentication using a third party identity provider is not " +
				"supported by this server.",
		}
	}
	oauth2State := crypto.NewToken(30)
	token := crypto.NewToken(256)
	now := time.Now().UTC()
	expiryTime := now.Add(s.config.UserSessionTTL)
	session := Session{
		ObjectMeta: meta.ObjectMeta{
			ID: uuid.NewV4().String(),
		},
		HashedOAuth2State: crypto.Hash("", oauth2State),
		HashedToken:       crypto.Hash("", token),
		Expires:           &expiryTime,
	}
	if opts != nil {
		session.AuthSuccessURL = opts.SuccessURL
	}
	session.Created = &now
	if err := s.sessionsStore.Create(ctx, session); err != nil {
		return ThirdPartyAuthDetails{}, errors.Wrapf(
			err,
			"error storing new user session %q",
			session.ID,
		)
	}
	return ThirdPartyAuthDetails{
		Token:   token,
		AuthURL: s.thirdPartyAuthHelper.AuthURL(oauth2State),
	}, nil
}

func (s *sessionsService) Authenticate(
	ctx context.Context,
	state string,
	code string,
) (string, error) {
	if !s.config.ThirdPartyAuthEnabled {
		return "", &meta.ErrNotSupported{
			Details: "Authentication using a third party identity provider is not " +
				"supported by this server.",
		}
	}
	session, err := s.sessionsStore.GetByHashedOAuth2State(
		ctx,
		crypto.Hash("", state),
	)
	if err != nil {
		return "", errors.Wrap(
			err,
			"error retrieving session from store by hashed OAuth2 state",
		)
	}
	thirdPartyUserIdentity, err :=
		s.thirdPartyAuthHelper.Exchange(ctx, state, code)
	if err != nil {
		return "",
			errors.Wrap(err, "error exchanging OAuth2 code for user identity")
	}
	user, err := s.usersStore.Get(ctx, thirdPartyUserIdentity.ID)
	if err != nil {
		if _, ok := errors.Cause(err).(*meta.ErrNotFound); ok {
			// User wasn't found. That's ok. We'll create one.
			user = User{
				ObjectMeta: meta.ObjectMeta{
					ID: thirdPartyUserIdentity.ID,
				},
				Name: thirdPartyUserIdentity.Name,
			}
			if err = s.usersStore.Create(ctx, user); err != nil {
				return "", errors.Wrapf(err, "error storing new user %q", user.ID)
			}
			for _, adminID := range s.config.AdminUserIDs {
				if user.ID == adminID {
					for _, role := range []libAuthz.Role{
						system.RoleAdmin,
						system.RoleProjectCreator,
						system.RoleReader,
					} {
						if err = s.grantRoleFn(
							ctx,
							libAuthz.RoleAssignment{
								Principal: libAuthz.PrincipalReference{
									Type: "USER",
									ID:   user.ID,
								},
								Role: role,
							},
						); err != nil {
							return "", errors.Wrapf(
								err,
								"error assigning role %q to user %q",
								role,
								user.ID,
							)
						}
					}
					break
				}
			}
		} else {
			// It was something else that went wrong when searching for the user.
			return "", err
		}
	}
	if err := s.sessionsStore.Authenticate(
		ctx,
		session.ID,
		user.ID,
		time.Now().UTC().Add(s.config.UserSessionTTL),
	); err != nil {
		return "", errors.Wrapf(
			err,
			"error storing authentication details for session %q",
			session.ID,
		)
	}
	return session.AuthSuccessURL, nil
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
	// of the OAuth 2 "state" code. This is used in completing both OpenID Connect
	// and GitHub authentication workflows. If no such Session exists,
	// implementations MUST return a *meta.ErrNotFound error.
	GetByHashedOAuth2State(context.Context, string) (Session, error)
	// GetByHashedToken returns a Session having the indicated secure hash of the
	// opaque bearer token. If no such Session exists, implementations MUST
	// return a *meta.ErrNotFound error.
	GetByHashedToken(context.Context, string) (Session, error)
	// Authenticate updates the specified, as-yet-anonymous Session (with an
	// as-yet unactivated token) to denote ownership by the indicated User and to
	// assign the specified expiry date/time. This is used in completing
	// third-party authentication workflows.
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
