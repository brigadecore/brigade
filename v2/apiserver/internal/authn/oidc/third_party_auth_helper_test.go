package oidc

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authn"
	"github.com/coreos/go-oidc"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestNewThirdPartyAuthHelper(t *testing.T) {
	oauth2Config := &mockOAuth2Config{}
	idTokenVerifier := &mockIDTokenVerifier{}
	helper := NewThirdPartyAuthHelper(oauth2Config, idTokenVerifier)
	require.Same(
		t,
		helper.(*thirdPartyAuthHelper).oauth2Config,
		oauth2Config,
	)
	require.Same(
		t,
		helper.(*thirdPartyAuthHelper).idTokenVerifier,
		idTokenVerifier,
	)
}

func TestAuthURL(t *testing.T) {
	const testState = "foo"
	testAuthURL := fmt.Sprintf("https://examples.com/auth?state=%s", testState)
	oauth2Config := &mockOAuth2Config{
		AuthCodeURLFn: func(string, ...oauth2.AuthCodeOption) string {
			return testAuthURL
		},
	}
	helper := NewThirdPartyAuthHelper(oauth2Config, nil)
	require.Equal(t, testAuthURL, helper.AuthURL(testState))
}

func TestExchange(t *testing.T) {
	const testOauth2State = "foo"
	const testOauth2Code = "bar"
	testCases := []struct {
		name       string
		helper     authn.ThirdPartyAuthHelper
		assertions func(err error)
	}{
		{
			name: "OAuth2 token exchange fails",
			helper: &thirdPartyAuthHelper{
				oauth2Config: &mockOAuth2Config{
					ExchangeFn: func(
						context.Context,
						string,
						...oauth2.AuthCodeOption,
					) (*oauth2.Token, error) {
						return nil, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					"error exchanging code for OAuth2 token",
				)
				require.Contains(t, err.Error(), "something went wrong")
			},
		},
		{
			name: "OAuth2 token does not contain an OpenID Connect identity token",
			helper: &thirdPartyAuthHelper{
				oauth2Config: &mockOAuth2Config{
					ExchangeFn: func(
						context.Context,
						string,
						...oauth2.AuthCodeOption,
					) (*oauth2.Token, error) {
						return &oauth2.Token{}, nil
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
			helper: &thirdPartyAuthHelper{
				oauth2Config: &mockOAuth2Config{
					ExchangeFn: func(
						context.Context,
						string,
						...oauth2.AuthCodeOption,
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
				idTokenVerifier: &mockIDTokenVerifier{
					VerifyFn: func(context.Context, string) (*oidc.IDToken, error) {
						return nil, errors.New("something went wrong")
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
				require.Contains(t, err.Error(), "something went wrong")
			},
		},
		{
			name: "success",
			helper: &thirdPartyAuthHelper{
				oauth2Config: &mockOAuth2Config{
					ExchangeFn: func(
						context.Context,
						string,
						...oauth2.AuthCodeOption,
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
				idTokenVerifier: &mockIDTokenVerifier{
					VerifyFn: func(context.Context, string) (*oidc.IDToken, error) {
						token := &oidc.IDToken{}
						setUnexportedField(
							token, "claims",
							[]byte(`{"name": "tony@starkindustries.com", "email": "tony@starkindustries.com"}`), // nolint: lll
						)
						return token, nil
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
			_, err := testCase.helper.Exchange(
				context.Background(),
				testOauth2State,
				testOauth2Code,
			)
			testCase.assertions(err)
		})
	}
}

type mockOAuth2Config struct {
	AuthCodeURLFn func(state string, opts ...oauth2.AuthCodeOption) string
	ExchangeFn    func(
		ctx context.Context,
		code string,
		opts ...oauth2.AuthCodeOption,
	) (*oauth2.Token, error)
}

func (m *mockOAuth2Config) AuthCodeURL(
	state string,
	opts ...oauth2.AuthCodeOption,
) string {
	return m.AuthCodeURLFn(state, opts...)
}

func (m *mockOAuth2Config) Exchange(
	ctx context.Context,
	code string,
	opts ...oauth2.AuthCodeOption,
) (*oauth2.Token, error) {
	return m.ExchangeFn(ctx, code, opts...)
}

type mockIDTokenVerifier struct {
	VerifyFn func(ctx context.Context, rawIDToken string) (*oidc.IDToken, error)
}

func (m *mockIDTokenVerifier) Verify(
	ctx context.Context,
	rawIDToken string,
) (*oidc.IDToken, error) {
	return m.VerifyFn(ctx, rawIDToken)
}

func setUnexportedField(
	objPtr interface{},
	fieldName string,
	fieldValue interface{},
) {
	field := reflect.ValueOf(objPtr).Elem().FieldByName(fieldName)
	reflect.NewAt(
		field.Type(),
		unsafe.Pointer(field.UnsafeAddr()),
	).Elem().Set(reflect.ValueOf(fieldValue))
}
