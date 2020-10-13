package authx

import (
	"context"
	"testing"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/crypto"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/coreos/go-oidc"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestOIDCAuthDetailsMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, OIDCAuthDetails{}, "OIDCAuthDetails")
}

func TestNewSessionsService(t *testing.T) {
	const testRootPassword = "12345"
	sessionsStore := &mockSessionsStore{}
	usersStore := &mockUsersStore{}
	config := &SessionsServiceConfig{
		RootUserEnabled:  true,
		RootUserPassword: testRootPassword,
	}
	svc := NewSessionsService(sessionsStore, usersStore, config)
	require.Same(t, sessionsStore, svc.(*sessionsService).sessionsStore)
	require.Same(t, usersStore, svc.(*sessionsService).usersStore)
	require.Equal(
		t,
		config.RootUserEnabled,
		svc.(*sessionsService).config.RootUserEnabled,
	)
	require.Equal(
		t,
		crypto.Hash("root", testRootPassword),
		svc.(*sessionsService).hashedRootUserPassword,
	)
	require.Empty(t, svc.(*sessionsService).config.RootUserPassword)
}

func TestSessionsServiceCreateRootSession(t *testing.T) {
	const testUsername = "root"
	const testPassword = "12345"

	testCases := []struct {
		name       string
		service    SessionsService
		username   string
		password   string
		assertions func(token Token, err error)
	}{
		{
			name: "root login not supported",
			service: &sessionsService{
				config: SessionsServiceConfig{
					RootUserEnabled: false,
				},
			},
			username: testUsername,
			password: testPassword,
			assertions: func(token Token, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotSupported{}, err)
			},
		},

		{
			name: "invalid credentials",
			service: &sessionsService{
				config: SessionsServiceConfig{
					RootUserEnabled: true,
				},
				hashedRootUserPassword: crypto.Hash(testUsername, testPassword),
			},
			username: testUsername,
			password: "WRONG!",
			assertions: func(token Token, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthentication{}, err)
			},
		},

		{
			name: "error creating session",
			service: &sessionsService{
				sessionsStore: &mockSessionsStore{
					CreateFn: func(context.Context, Session) error {
						return errors.New("error creating new session")
					},
				},
				config: SessionsServiceConfig{
					RootUserEnabled: true,
				},
				hashedRootUserPassword: crypto.Hash(testUsername, testPassword),
			},
			username: testUsername,
			password: testPassword,
			assertions: func(token Token, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error storing new root session")
				require.Contains(t, err.Error(), "error creating new session")
			},
		},

		{
			name: "success",
			service: &sessionsService{
				sessionsStore: &mockSessionsStore{
					CreateFn: func(context.Context, Session) error {
						return nil
					},
				},
				config: SessionsServiceConfig{
					RootUserEnabled: true,
				},
				hashedRootUserPassword: crypto.Hash(testUsername, testPassword),
			},
			username: testUsername,
			password: testPassword,
			assertions: func(token Token, err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			token, err := testCase.service.CreateRootSession(
				context.Background(),
				testCase.username,
				testCase.password,
			)
			testCase.assertions(token, err)
		})
	}
}

func TestSessionsServiceCreateUserSession(t *testing.T) {
	const testAuthURL = "https://localhost:8080/oidc?state=foo"
	testCases := []struct {
		name       string
		service    SessionsService
		assertions func(authDetails OIDCAuthDetails, err error)
	}{
		{
			name: "OpenID Connect not supported",
			service: &sessionsService{
				config: SessionsServiceConfig{
					OpenIDConnectEnabled: false,
				},
			},
			assertions: func(authDetails OIDCAuthDetails, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotSupported{}, err)
			},
		},
		{
			name: "OpenID Connect supported",
			service: &sessionsService{
				sessionsStore: &mockSessionsStore{
					CreateFn: func(context.Context, Session) error {
						return nil
					},
				},
				config: SessionsServiceConfig{
					OpenIDConnectEnabled: true,
					OAuth2Helper: &mockOAuth2Helper{
						AuthCodeURLFn: func(
							state string,
							opts ...oauth2.AuthCodeOption,
						) string {
							return testAuthURL
						},
					},
				},
			},
			assertions: func(authDetails OIDCAuthDetails, err error) {
				require.NoError(t, err)
				require.Len(t, authDetails.Token, 256)
				require.Equal(t, testAuthURL, authDetails.AuthURL)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			authDetails, err :=
				testCase.service.CreateUserSession(context.Background())
			testCase.assertions(authDetails, err)
		})
	}
}

func TestSessionsServiceAuthenticate(t *testing.T) {
	testCases := []struct {
		name        string
		oauth2State string
		oidcCode    string
		service     SessionsService
		assertions  func(err error)
	}{

		{
			name: "OpenID Connect not supported",
			service: &sessionsService{
				config: SessionsServiceConfig{
					OpenIDConnectEnabled: false,
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotSupported{}, err)
			},
		},

		{
			name: "OAuth2 state not found",
			service: &sessionsService{
				sessionsStore: &mockSessionsStore{
					GetByHashedOAuth2StateFn: func(
						ctx context.Context,
						oauth2State string,
					) (Session, error) {
						return Session{}, &meta.ErrNotFound{
							Type: "Session",
						}
					},
				},
				config: SessionsServiceConfig{
					OpenIDConnectEnabled: true,
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, errors.Cause(err))
				require.Equal(t, "Session", errors.Cause(err).(*meta.ErrNotFound).Type)
			},
		},

		{
			name: "OAuth2 token exchange fails",
			service: &sessionsService{
				sessionsStore: &mockSessionsStore{
					GetByHashedOAuth2StateFn: func(
						ctx context.Context,
						oauth2State string,
					) (Session, error) {
						return Session{}, nil
					},
				},
				config: SessionsServiceConfig{
					OpenIDConnectEnabled: true,
					OAuth2Helper: &mockOAuth2Helper{
						ExchangeFn: func(
							ctx context.Context,
							code string,
							opts ...oauth2.AuthCodeOption,
						) (*oauth2.Token, error) {
							return nil, errors.New("error exchanging token")
						},
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					"error exchanging OpenID Connect code for OAuth2 token",
				)
				require.Contains(t, err.Error(), "error exchanging token")
			},
		},

		{
			name: "OAuth2 token does not contain an OpenID Connect identity token",
			service: &sessionsService{
				sessionsStore: &mockSessionsStore{
					GetByHashedOAuth2StateFn: func(
						ctx context.Context,
						oauth2State string,
					) (Session, error) {
						return Session{}, nil
					},
				},
				config: SessionsServiceConfig{
					OpenIDConnectEnabled: true,
					OAuth2Helper: &mockOAuth2Helper{
						ExchangeFn: func(
							ctx context.Context,
							code string,
							opts ...oauth2.AuthCodeOption,
						) (*oauth2.Token, error) {
							return &oauth2.Token{}, nil
						},
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					"did not include an OpenID Connect identity token",
				)
			},
		},

		{
			name: "error verifying OpenID Connect identity token",
			service: &sessionsService{
				sessionsStore: &mockSessionsStore{
					GetByHashedOAuth2StateFn: func(
						ctx context.Context,
						oauth2State string,
					) (Session, error) {
						return Session{}, nil
					},
				},
				config: SessionsServiceConfig{
					OpenIDConnectEnabled: true,
					OAuth2Helper: &mockOAuth2Helper{
						ExchangeFn: func(
							ctx context.Context,
							code string,
							opts ...oauth2.AuthCodeOption,
						) (*oauth2.Token, error) {
							token := &oauth2.Token{}
							setUnexportedField(
								token,
								"raw",
								map[string]interface{}{
									"id_token": "fakeidtoken",
								},
							)
							return token, nil
						},
					},
					OpenIDConnectTokenVerifier: &mockOpenIDConnectTokenVerifier{
						VerifyFn: func(
							ctx context.Context,
							rawIDToken string,
						) (*oidc.IDToken, error) {
							return nil, errors.New("error verifying token")
						},
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					"error verifying OpenID Connect identity token",
				)
				require.Contains(
					t,
					err.Error(),
					"error verifying token",
				)
			},
		},

		{
			name: "error finding existing user",
			service: &sessionsService{
				sessionsStore: &mockSessionsStore{
					GetByHashedOAuth2StateFn: func(
						ctx context.Context,
						oauth2State string,
					) (Session, error) {
						return Session{}, nil
					},
				},
				usersStore: &mockUsersStore{
					GetFn: func(_ context.Context, id string) (User, error) {
						return User{}, errors.New("error searching for user")
					},
				},
				config: SessionsServiceConfig{
					OpenIDConnectEnabled: true,
					OAuth2Helper: &mockOAuth2Helper{
						ExchangeFn: func(
							ctx context.Context,
							code string,
							opts ...oauth2.AuthCodeOption,
						) (*oauth2.Token, error) {
							token := &oauth2.Token{}
							setUnexportedField(
								token,
								"raw",
								map[string]interface{}{
									"id_token": "fakeidtoken",
								},
							)
							return token, nil
						},
					},
					OpenIDConnectTokenVerifier: &mockOpenIDConnectTokenVerifier{
						VerifyFn: func(
							ctx context.Context,
							rawIDToken string,
						) (*oidc.IDToken, error) {
							token := &oidc.IDToken{}
							setUnexportedField(
								token, "claims",
								[]byte(`{"name": "tony@starkindustries.com", "email": "tony@starkindustries.com"}`), // nolint: lll
							)
							return token, nil
						},
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error searching for user")
			},
		},

		{
			name: "existing user not found; error creating user",
			service: &sessionsService{
				sessionsStore: &mockSessionsStore{
					GetByHashedOAuth2StateFn: func(
						ctx context.Context,
						oauth2State string,
					) (Session, error) {
						return Session{}, nil
					},
				},
				usersStore: &mockUsersStore{
					GetFn: func(_ context.Context, id string) (User, error) {
						return User{}, &meta.ErrNotFound{
							Type: "User",
							ID:   id,
						}
					},
					CreateFn: func(context.Context, User) error {
						return errors.New("error creating new user")
					},
				},
				config: SessionsServiceConfig{
					OpenIDConnectEnabled: true,
					OAuth2Helper: &mockOAuth2Helper{
						ExchangeFn: func(
							ctx context.Context,
							code string,
							opts ...oauth2.AuthCodeOption,
						) (*oauth2.Token, error) {
							token := &oauth2.Token{}
							setUnexportedField(
								token,
								"raw",
								map[string]interface{}{
									"id_token": "fakeidtoken",
								},
							)
							return token, nil
						},
					},
					OpenIDConnectTokenVerifier: &mockOpenIDConnectTokenVerifier{
						VerifyFn: func(
							ctx context.Context,
							rawIDToken string,
						) (*oidc.IDToken, error) {
							token := &oidc.IDToken{}
							setUnexportedField(
								token, "claims",
								[]byte(`{"name": "tony@starkindustries.com", "email": "tony@starkindustries.com"}`), // nolint: lll
							)
							return token, nil
						},
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error storing new user")
				require.Contains(t, err.Error(), "error creating new user")
			},
		},

		{
			name: "error authenticating in session store",
			service: &sessionsService{
				sessionsStore: &mockSessionsStore{
					GetByHashedOAuth2StateFn: func(
						ctx context.Context,
						oauth2State string,
					) (Session, error) {
						return Session{}, nil
					},
					AuthenticateFn: func(
						ctx context.Context,
						sessionID string,
						userID string,
						expires time.Time,
					) error {
						return errors.New("error authenticating")
					},
				},
				usersStore: &mockUsersStore{
					GetFn: func(_ context.Context, id string) (User, error) {
						return User{}, &meta.ErrNotFound{
							Type: "User",
							ID:   id,
						}
					},
					CreateFn: func(context.Context, User) error {
						return nil
					},
				},
				config: SessionsServiceConfig{
					OpenIDConnectEnabled: true,
					OAuth2Helper: &mockOAuth2Helper{
						ExchangeFn: func(
							ctx context.Context,
							code string,
							opts ...oauth2.AuthCodeOption,
						) (*oauth2.Token, error) {
							token := &oauth2.Token{}
							setUnexportedField(
								token,
								"raw",
								map[string]interface{}{
									"id_token": "fakeidtoken",
								},
							)
							return token, nil
						},
					},
					OpenIDConnectTokenVerifier: &mockOpenIDConnectTokenVerifier{
						VerifyFn: func(
							ctx context.Context,
							rawIDToken string,
						) (*oidc.IDToken, error) {
							token := &oidc.IDToken{}
							setUnexportedField(
								token, "claims",
								[]byte(`{"name": "tony@starkindustries.com", "email": "tony@starkindustries.com"}`), // nolint: lll
							)
							return token, nil
						},
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					"error storing authentication details for session",
				)
				require.Contains(t, err.Error(), "error authenticating")
			},
		},

		{
			name: "success",
			service: &sessionsService{
				sessionsStore: &mockSessionsStore{
					GetByHashedOAuth2StateFn: func(
						ctx context.Context,
						oauth2State string,
					) (Session, error) {
						return Session{}, nil
					},
					AuthenticateFn: func(
						ctx context.Context,
						sessionID string,
						userID string,
						expires time.Time,
					) error {
						return nil
					},
				},
				usersStore: &mockUsersStore{
					GetFn: func(_ context.Context, id string) (User, error) {
						return User{}, nil
					},
					CreateFn: func(context.Context, User) error {
						return nil
					},
				},
				config: SessionsServiceConfig{
					OpenIDConnectEnabled: true,
					OAuth2Helper: &mockOAuth2Helper{
						ExchangeFn: func(
							ctx context.Context,
							code string,
							opts ...oauth2.AuthCodeOption,
						) (*oauth2.Token, error) {
							token := &oauth2.Token{}
							setUnexportedField(
								token,
								"raw",
								map[string]interface{}{
									"id_token": "fakeidtoken",
								},
							)
							return token, nil
						},
					},
					OpenIDConnectTokenVerifier: &mockOpenIDConnectTokenVerifier{
						VerifyFn: func(
							ctx context.Context,
							rawIDToken string,
						) (*oidc.IDToken, error) {
							token := &oidc.IDToken{}
							setUnexportedField(
								token, "claims",
								[]byte(`{"name": "tony@starkindustries.com", "email": "tony@starkindustries.com"}`), // nolint: lll
							)
							return token, nil
						},
					},
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.service.Authenticate(
				context.Background(),
				testCase.oauth2State,
				testCase.oidcCode,
			)
			testCase.assertions(err)
		})
	}
}

func TestSessionsServiceGetByToken(t *testing.T) {
	const testSessionID = "12345"
	testCases := []struct {
		name       string
		service    SessionsService
		assertions func(session Session, err error)
	}{
		{
			name: "session not found",
			service: &sessionsService{
				sessionsStore: &mockSessionsStore{
					GetByHashedTokenFn: func(context.Context, string) (Session, error) {
						return Session{}, &meta.ErrNotFound{
							Type: "Session",
						}
					},
				},
			},
			assertions: func(session Session, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, errors.Cause(err))
				require.Equal(t, "Session", errors.Cause(err).(*meta.ErrNotFound).Type)
			},
		},
		{
			name: "session found",
			service: &sessionsService{
				sessionsStore: &mockSessionsStore{
					GetByHashedTokenFn: func(context.Context, string) (Session, error) {
						return Session{
							ObjectMeta: meta.ObjectMeta{
								ID: testSessionID,
							},
						}, nil
					},
				},
			},
			assertions: func(session Session, err error) {
				require.NoError(t, err)
				require.Equal(t, testSessionID, session.ID)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			session, err :=
				testCase.service.GetByToken(context.Background(), "faketoken")
			testCase.assertions(session, err)
		})
	}
}

func TestSessionsServiceDelete(t *testing.T) {
	const testSessionID = "12345"
	testCases := []struct {
		name       string
		service    SessionsService
		assertions func(err error)
	}{
		{
			name: "session not found",
			service: &sessionsService{
				sessionsStore: &mockSessionsStore{
					DeleteFn: func(context.Context, string) error {
						return &meta.ErrNotFound{
							Type: "Session",
						}
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, errors.Cause(err))
				require.Equal(t, "Session", errors.Cause(err).(*meta.ErrNotFound).Type)
			},
		},
		{
			name: "session found",
			service: &sessionsService{
				sessionsStore: &mockSessionsStore{
					DeleteFn: func(context.Context, string) error {
						return nil
					},
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.service.Delete(context.Background(), testSessionID)
			testCase.assertions(err)
		})
	}
}

type mockSessionsStore struct {
	CreateFn                 func(context.Context, Session) error
	GetByHashedOAuth2StateFn func(context.Context, string) (Session, error)
	GetByHashedTokenFn       func(context.Context, string) (Session, error)
	AuthenticateFn           func(
		ctx context.Context,
		sessionID string,
		userID string,
		expires time.Time,
	) error
	DeleteFn func(context.Context, string) error
}

func (m *mockSessionsStore) Create(ctx context.Context, session Session) error {
	return m.CreateFn(ctx, session)
}

func (m *mockSessionsStore) GetByHashedOAuth2State(
	ctx context.Context,
	hashedOAuth2State string,
) (Session, error) {
	return m.GetByHashedOAuth2StateFn(ctx, hashedOAuth2State)
}

func (m *mockSessionsStore) GetByHashedToken(
	ctx context.Context,
	hashedToken string,
) (Session, error) {
	return m.GetByHashedTokenFn(ctx, hashedToken)
}

func (m *mockSessionsStore) Authenticate(
	ctx context.Context,
	sessionID string,
	userID string,
	expires time.Time,
) error {
	return m.AuthenticateFn(ctx, sessionID, userID, expires)
}

func (m *mockSessionsStore) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

type mockOAuth2Helper struct {
	AuthCodeURLFn func(state string, opts ...oauth2.AuthCodeOption) string
	ExchangeFn    func(
		ctx context.Context,
		code string,
		opts ...oauth2.AuthCodeOption,
	) (*oauth2.Token, error)
}

func (m *mockOAuth2Helper) AuthCodeURL(
	state string,
	opts ...oauth2.AuthCodeOption,
) string {
	return m.AuthCodeURLFn(state, opts...)
}

func (m *mockOAuth2Helper) Exchange(
	ctx context.Context,
	code string,
	opts ...oauth2.AuthCodeOption,
) (*oauth2.Token, error) {
	return m.ExchangeFn(ctx, code, opts...)
}

type mockOpenIDConnectTokenVerifier struct {
	VerifyFn func(ctx context.Context, rawIDToken string) (*oidc.IDToken, error)
}

func (m *mockOpenIDConnectTokenVerifier) Verify(
	ctx context.Context,
	rawIDToken string,
) (*oidc.IDToken, error) {
	return m.VerifyFn(ctx, rawIDToken)
}
