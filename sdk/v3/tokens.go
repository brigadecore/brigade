package sdk

// Token represents an opaque bearer token used to authenticate to the Brigade
// API.
type Token struct {
	Value string `json:"value,omitempty"`
}
