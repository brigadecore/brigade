package restmachinery

// APIClientOptions encapsulates optional API client configuration.
type APIClientOptions struct {
	// AllowInsecureConnections indicates whether SSL-related errors should be
	// ignored when connecting to the API server.
	AllowInsecureConnections bool
}
