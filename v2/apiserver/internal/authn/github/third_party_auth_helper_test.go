package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authn"
	"github.com/stretchr/testify/require"
)

func TestNewThirdPartyAuthHelper(t *testing.T) {
	config := ThirdPartyAuthHelperConfig{}
	helper := NewThirdPartyAuthHelper(config)
	require.Equal(t, helper.(*thirdPartyAuthHelper).config, config)
}

func TestAuthURL(t *testing.T) {
	const testClientID = "foo"
	const testState = "foo"
	testAuthURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&state=%s&scope=read%%3Aorg", // nolint: lll
		testClientID,
		testState,
	)
	helper := NewThirdPartyAuthHelper(
		ThirdPartyAuthHelperConfig{
			ClientID: testClientID,
		},
	)
	require.Equal(t, testAuthURL, helper.AuthURL(testState))
}

func TestExchangeURL(t *testing.T) {
	const testOauth2State = "foo"
	const testOauth2Code = "bar"
	const testAccessToken = "opensesame"
	testCases := []struct {
		name                  string
		helper                authn.ThirdPartyAuthHelper
		mockGitHubHandlerFunc http.HandlerFunc
		assertions            func(err error)
	}{
		{
			name: "error exchanging code for OAuth2 access token",
			helper: &thirdPartyAuthHelper{
				getOAuth2AccessTokenFn: func(string, string) (string, error) {
					return "", errors.New("something went wrong")
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					"error exchanging code for OAuth2 access token",
				)
				require.Contains(t, err.Error(), "something went wrong")
			},
		},
		{
			name: "error retrieving user identity from GitHub",
			helper: &thirdPartyAuthHelper{
				getOAuth2AccessTokenFn: func(string, string) (string, error) {
					return testAccessToken, nil
				},
				getUserIdentityFn: func(
					context.Context,
					string,
				) (authn.ThirdPartyIdentity, error) {
					return authn.ThirdPartyIdentity{}, errors.New("something went wrong")
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					"error retrieving user identity from GitHub",
				)
				require.Contains(t, err.Error(), "something went wrong")
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
