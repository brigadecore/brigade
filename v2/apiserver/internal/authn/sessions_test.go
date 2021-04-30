package authn

import (
	"context"
	"testing"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/crypto"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	metaTesting "github.com/brigadecore/brigade/v2/apiserver/internal/meta/testing" // nolint: lll
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestThirdPartyAuthDetailsMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(
		t,
		ThirdPartyAuthDetails{},
		"ThirdPartyAuthDetails",
	)
}

func TestNewSessionsService(t *testing.T) {
	const testRootPassword = "12345"
	sessionsStore := &mockSessionsStore{}
	usersStore := &MockUsersStore{}
	thirdPartyAuthHelper := &MockThirdPartyAuthHelper{}
	config := &SessionsServiceConfig{
		RootUserEnabled:  true,
		RootUserPassword: testRootPassword,
	}
	svc := NewSessionsService(
		sessionsStore,
		usersStore,
		nil,
		thirdPartyAuthHelper,
		config,
	)
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
	testThirdPartyAuthOpts := &ThirdPartyAuthOptions{
		SuccessURL: "https://example.com/success",
	}
	testCases := []struct {
		name       string
		service    SessionsService
		assertions func(authDetails ThirdPartyAuthDetails, err error)
	}{
		{
			name: "third-party authentication disabled",
			service: &sessionsService{
				config: SessionsServiceConfig{
					ThirdPartyAuthEnabled: false,
				},
			},
			assertions: func(authDetails ThirdPartyAuthDetails, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotSupported{}, err)
			},
		},
		{
			name: "third-party authentication enabled",
			service: &sessionsService{
				sessionsStore: &mockSessionsStore{
					CreateFn: func(context.Context, Session) error {
						return nil
					},
				},
				thirdPartyAuthHelper: &MockThirdPartyAuthHelper{
					AuthURLFn: func(oauth2State string) string {
						return testAuthURL
					},
				},
				config: SessionsServiceConfig{
					ThirdPartyAuthEnabled: true,
				},
			},
			assertions: func(authDetails ThirdPartyAuthDetails, err error) {
				require.NoError(t, err)
				require.Len(t, authDetails.Token, 256)
				require.Equal(t, testAuthURL, authDetails.AuthURL)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			authDetails, err := testCase.service.CreateUserSession(
				context.Background(),
				testThirdPartyAuthOpts,
			)
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
			name: "third-party authentication disabled",
			service: &sessionsService{
				config: SessionsServiceConfig{
					ThirdPartyAuthEnabled: false,
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
							Type: SessionKind,
						}
					},
				},
				config: SessionsServiceConfig{
					ThirdPartyAuthEnabled: true,
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, errors.Cause(err))
				require.Equal(
					t,
					SessionKind,
					errors.Cause(err).(*meta.ErrNotFound).Type,
				)
			},
		},

		{
			name: "OAuth2 code/identity exchange fails",
			service: &sessionsService{
				sessionsStore: &mockSessionsStore{
					GetByHashedOAuth2StateFn: func(
						ctx context.Context,
						oauth2State string,
					) (Session, error) {
						return Session{}, nil
					},
				},
				thirdPartyAuthHelper: &MockThirdPartyAuthHelper{
					ExchangeFn: func(
						context.Context,
						string,
						string,
					) (ThirdPartyIdentity, error) {
						return ThirdPartyIdentity{}, errors.New("something went wrong")
					},
				},
				config: SessionsServiceConfig{
					ThirdPartyAuthEnabled: true,
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					"error exchanging OAuth2 code for user identity",
				)
				require.Contains(t, err.Error(), "something went wrong")
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
				usersStore: &MockUsersStore{
					GetFn: func(_ context.Context, id string) (User, error) {
						return User{}, errors.New("something went wrong")
					},
				},
				thirdPartyAuthHelper: &MockThirdPartyAuthHelper{
					ExchangeFn: func(
						context.Context,
						string,
						string,
					) (ThirdPartyIdentity, error) {
						return ThirdPartyIdentity{}, nil
					},
				},
				config: SessionsServiceConfig{
					ThirdPartyAuthEnabled: true,
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
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
				usersStore: &MockUsersStore{
					GetFn: func(_ context.Context, id string) (User, error) {
						return User{}, &meta.ErrNotFound{
							Type: UserKind,
							ID:   id,
						}
					},
					CreateFn: func(context.Context, User) error {
						return errors.New("something went wrong")
					},
				},
				thirdPartyAuthHelper: &MockThirdPartyAuthHelper{
					ExchangeFn: func(
						context.Context,
						string,
						string,
					) (ThirdPartyIdentity, error) {
						return ThirdPartyIdentity{}, nil
					},
				},
				config: SessionsServiceConfig{
					ThirdPartyAuthEnabled: true,
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error storing new user")
				require.Contains(t, err.Error(), "something went wrong")
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
						return errors.New("something went wrong")
					},
				},
				usersStore: &MockUsersStore{
					GetFn: func(_ context.Context, id string) (User, error) {
						return User{}, &meta.ErrNotFound{
							Type: UserKind,
							ID:   id,
						}
					},
					CreateFn: func(context.Context, User) error {
						return nil
					},
				},
				thirdPartyAuthHelper: &MockThirdPartyAuthHelper{
					ExchangeFn: func(
						context.Context,
						string,
						string,
					) (ThirdPartyIdentity, error) {
						return ThirdPartyIdentity{}, nil
					},
				},
				config: SessionsServiceConfig{
					ThirdPartyAuthEnabled: true,
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					"error storing authentication details for session",
				)
				require.Contains(t, err.Error(), "something went wrong")
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
				usersStore: &MockUsersStore{
					GetFn: func(_ context.Context, id string) (User, error) {
						return User{}, nil
					},
					CreateFn: func(context.Context, User) error {
						return nil
					},
				},
				thirdPartyAuthHelper: &MockThirdPartyAuthHelper{
					ExchangeFn: func(
						context.Context,
						string,
						string,
					) (ThirdPartyIdentity, error) {
						return ThirdPartyIdentity{}, nil
					},
				},
				config: SessionsServiceConfig{
					ThirdPartyAuthEnabled: true,
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := testCase.service.Authenticate(
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
							Type: SessionKind,
						}
					},
				},
			},
			assertions: func(session Session, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, errors.Cause(err))
				require.Equal(
					t,
					SessionKind,
					errors.Cause(err).(*meta.ErrNotFound).Type,
				)
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
							Type: SessionKind,
						}
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, errors.Cause(err))
				require.Equal(
					t,
					SessionKind,
					errors.Cause(err).(*meta.ErrNotFound).Type,
				)
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
	DeleteFn       func(context.Context, string) error
	DeleteByUserFn func(context.Context, string) error
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

func (m *mockSessionsStore) DeleteByUser(
	ctx context.Context,
	userID string,
) error {
	return m.DeleteByUserFn(ctx, userID)
}
