package oidc

import (
	"context"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authn"
	"github.com/coreos/go-oidc"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// OAuth2Config is an interface for the subset of *oauth2.Config functions used
// for Brigade Session management. Dependence on this interface instead of
// directly upon the *oauth2.Config allows for the possibility of utilizing a
// mock implementation for testing purposes. Adding only the subset of functions
// that we actually use limits the effort involved in creating such mocks.
type OAuth2Config interface {
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

// IDTokenVerifier is an interface for the subset of *oidc.IDTokenVerifier used
// for Brigade Session management. Dependence on this interface instead of
// directly upon the *oidc.IDTokenVerifier allows for the possibility of
// utilizing a mock implementation for testing purposes. Adding only the subset
// of functions that we actually use limits the effort involved in creating such
// mocks.
type IDTokenVerifier interface {
	Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error)
}

type thirdPartyAuthHelper struct {
	oauth2Config    OAuth2Config
	idTokenVerifier IDTokenVerifier
}

func NewThirdPartyAuthHelper(
	oauth2Config OAuth2Config,
	idTokenVerifier IDTokenVerifier,
) authn.ThirdPartyAuthHelper {
	return &thirdPartyAuthHelper{
		oauth2Config:    oauth2Config,
		idTokenVerifier: idTokenVerifier,
	}
}

func (t *thirdPartyAuthHelper) AuthURL(oauth2State string) string {
	return t.oauth2Config.AuthCodeURL(oauth2State)
}

func (t *thirdPartyAuthHelper) Exchange(
	ctx context.Context,
	_ string,
	oauth2Code string,
) (authn.ThirdPartyIdentity, error) {
	identity := authn.ThirdPartyIdentity{}
	oauth2Token, err := t.oauth2Config.Exchange(ctx, oauth2Code)
	if err != nil {
		return identity, errors.Wrap(err, "error exchanging code for OAuth2 token")
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return identity, errors.New(
			"OAuth2 token did not include an OpenID Connect identity token",
		)
	}
	var idToken *oidc.IDToken
	if idToken, err =
		t.idTokenVerifier.Verify(ctx, rawIDToken); err != nil {
		return identity,
			errors.Wrap(err, "error verifying OpenID Connect identity token")
	}
	claims := struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}{}
	if err = idToken.Claims(&claims); err != nil {
		return identity, errors.Wrap(
			err,
			"error decoding OpenID Connect identity token claims",
		)
	}
	identity.ID = claims.Email
	identity.Name = claims.Name
	return identity, nil
}
