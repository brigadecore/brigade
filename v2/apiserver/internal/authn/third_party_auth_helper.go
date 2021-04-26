package authn

import "context"

// ThirdPartyIdentity encapsulates ID (handle or email address) and name
// information for a User obtained from a third-party identity provider.
type ThirdPartyIdentity struct {
	// ID is a handle or email address for the User.
	ID string
	// Name is the User's given name + surname.
	Name string
}

// ThirdPartyAuthHelper is an interface for components that implement pluggable
// portions of third party authentication schemes based on OAuth2. OpenID
// Connect (used by certain identity providers like Azure Active Directory or
// Google Cloud Identity Platform) is built on top of OAuth2 (strictly speaking
// OAuth2 is for authorization; not authentication, but OpenID Connect extends
// OAuth2 with an authentication standard), but other authentication schemes
// (like GitHub's) are ALSO based on OAuth2, but DON'T implement OpenID Connect.
// This interface allows universal parts of authentication based on OAuth2 to be
// common for all third-party identity providers while utilizing
// standard-specific or provider-specific functionality for the portions of
// those authentication schemes that vary.
type ThirdPartyAuthHelper interface {
	// AuthURL returns a URL to which a User may be redirected for purposes of
	// completing authentication using a third-party identity provider.
	AuthURL(oauth2State string) string
	// Exchange exchanges a code issued by a third-party identity provider for
	// User identity information.
	Exchange(
		ctx context.Context,
		oauth2State string,
		oauth2Code string,
	) (ThirdPartyIdentity, error)
}

type MockThirdPartyAuthHelper struct {
	AuthURLFn  func(oauth2State string) string
	ExchangeFn func(
		ctx context.Context,
		oauth2State string,
		oauth2Code string,
	) (ThirdPartyIdentity, error)
}

func (m *MockThirdPartyAuthHelper) AuthURL(oauth2State string) string {
	return m.AuthURLFn(oauth2State)
}

func (m *MockThirdPartyAuthHelper) Exchange(
	ctx context.Context,
	oauth2State string,
	oauth2Code string,
) (ThirdPartyIdentity, error) {
	return m.ExchangeFn(ctx, oauth2State, oauth2Code)
}
