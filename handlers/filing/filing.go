package filing

import (
	"net/http"
)

// Streaming represents the configuration for HTTP handlers that output stream data.
type Streaming struct {
}

//Transform versioning functionality is not required for existing streams
func (st Streaming) Transform(req *http.Request, msg *[]byte) (err error) {
	return nil
}

//Filter functionality is not required for existing streams
//TODO - This method will filtor out sensitive data from the stream.
func (st Streaming) Filter(req *http.Request, msg *[]byte) error {
	return nil
}
