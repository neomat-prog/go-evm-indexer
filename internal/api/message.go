package api

import (
	"net/http"
)

type MessageResponse struct {
	Message string `json:"message"`
}

func (s *Server) handleHello(w http.ResponseWriter, r *http.Request) error {
	return writeJSON(w, http.StatusOK, MessageResponse{Message: "Hello, World!"})
}
