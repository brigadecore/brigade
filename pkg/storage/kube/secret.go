package kube

// SecretValues provides accessor methods for secrets.
type SecretValues map[string][]byte

// Bytes returns the value in the map for the provided key.
func (sv SecretValues) Bytes(key string) []byte {
	return sv[key]
}

// Bytes returns the string value in the map for the provided key.
func (sv SecretValues) String(key string) string {
	return string(sv.Bytes(key))
}
