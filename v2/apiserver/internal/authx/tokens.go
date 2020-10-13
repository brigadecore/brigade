package authx

import (
	"encoding/json"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
)

// Token represents an opaque bearer token used to authenticate to the Brigade
// API.
type Token struct {
	Value string `json:"value" bson:"value"`
}

// MarshalJSON amends Token instances with type metadata.
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
