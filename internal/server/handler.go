package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type apiFunc func(http.ResponseWriter, *http.Request) error

type apiError struct {
	Err    string `json:"error"`
	Status int    `json:"-"`
}

func (e apiError) Error() string { return e.Err }

// makeHTTPHandleFunc adapts an error-returning handler into a stdlib handler.
func makeHTTPHandleFunc(fn apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)
		if err == nil {
			return
		}

		var apiErr apiError
		if !errors.As(err, &apiErr) {
			log.Printf("error: %s %s: %v", r.Method, r.URL.Path, err)
			apiErr = apiError{Err: "internal server error", Status: http.StatusInternalServerError}
		}
		if apiErr.Status == 0 {
			apiErr.Status = http.StatusInternalServerError
		}
		writeJSON(w, apiErr.Status, apiErr)
	}
}

// writeJSON marshals before writing, so a failure returns before any header is
// sent and the error path can still respond. Write errors mean the client is
// gone; nothing to report.
func writeJSON(w http.ResponseWriter, status int, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
	return nil
}
