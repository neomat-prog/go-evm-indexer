package server

import "net/http"

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", makeHTTPHandleFunc(s.handleHealth))
	mux.HandleFunc("GET /logs", makeHTTPHandleFunc(s.handleTokenLog))
	mux.HandleFunc("GET /index", makeHTTPHandleFunc(s.handleIndexerState))
	mux.HandleFunc("GET /transfers", makeHTTPHandleFunc(s.handleLatestTransfers))
	return mux
}
