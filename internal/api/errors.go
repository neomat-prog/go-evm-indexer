package api

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

func newAPIError(status int, msg string) apiError {
	return apiError{Err: msg, Status: status}
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// trackingWriter remembers whether the handler already committed a status, so
// the error path can't emit a second one ("superfluous WriteHeader" + a body
// glued onto a successful response).
type trackingWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (w *trackingWriter) WriteHeader(status int) {
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *trackingWriter) Write(b []byte) (int, error) {
	w.wroteHeader = true
	return w.ResponseWriter.Write(b)
}

// makeHTTPHandleFunc adapts an error-returning handler into a stdlib handler.
func makeHTTPHandleFunc(fn apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tw := &trackingWriter{ResponseWriter: w}

		err := fn(tw, r)
		if err == nil {
			return
		}

		status := http.StatusInternalServerError
		payload := apiError{Err: "internal server error"}

		var apiErr apiError
		if errors.As(err, &apiErr) {
			payload = apiErr
			if apiErr.Status != 0 {
				status = apiErr.Status
			}
		}

		if status >= http.StatusInternalServerError {
			log.Printf("error: %s %s: %v", r.Method, r.URL.Path, err)
		}

		// Response already on the wire; nothing left to do but log it.
		if tw.wroteHeader {
			return
		}

		if err := writeJSON(tw, status, payload); err != nil {
			log.Printf("write error: %s %s: %v", r.Method, r.URL.Path, err)
		}
	}
}
