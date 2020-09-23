package authx

import (
	"encoding/json"

	"github.com/brigadecore/brigade/sdk/v2/meta"
)

// Token represents an opaque bearer token used to authenticate to the Brigade
// API.
type Token struct {
	Value string `json:"value,omitempty"`
}

// MarshalJSON amends Token instances with type metadata so that
// clients do not need to be concerned with the tedium of doing so.
func (t Token) MarshalJSON() ([]byte, error) {
	type Alias Token
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "Token",
			},
			Alias: (Alias)(t),
		},
	)
}
