package api

import "context"

type mockThirdPartyAuthHelper struct {
	AuthURLFn  func(oauth2State string) string
	ExchangeFn func(
		ctx context.Context,
		oauth2State string,
		oauth2Code string,
	) (ThirdPartyIdentity, error)
}

func (m *mockThirdPartyAuthHelper) AuthURL(oauth2State string) string {
	return m.AuthURLFn(oauth2State)
}

func (m *mockThirdPartyAuthHelper) Exchange(
	ctx context.Context,
	oauth2State string,
	oauth2Code string,
) (ThirdPartyIdentity, error) {
	return m.ExchangeFn(ctx, oauth2State, oauth2Code)
}
