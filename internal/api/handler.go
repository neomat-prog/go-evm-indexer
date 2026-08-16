package api

import "net/http"

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) error {
	return writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
