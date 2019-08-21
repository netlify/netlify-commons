package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// SendJSON will write the response object as JSON
func SendJSON(w http.ResponseWriter, status int, obj interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	b, err := json.Marshal(obj)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error encoding json response: %v", obj))
	}
	w.WriteHeader(status)
	_, err = w.Write(b)
	return err
}

// HandlerStatusOK is a APIHandler that just returns a 200 OK.
func HandlerStatusOK(w http.ResponseWriter, _ *http.Request) *HTTPError {
	w.WriteHeader(http.StatusOK)
	return nil
}
