package ports

// APIClient is the outgoing port for the RavenPair REST API.
// Implementations return the raw HTTP status code and response body so the
// caller can decide how to format and present the data.
type APIClient interface {
	GetStatus() (statusCode int, body []byte, err error)
	ListPairs() (statusCode int, body []byte, err error)
	CreatePair(name string) (statusCode int, body []byte, err error)
}
