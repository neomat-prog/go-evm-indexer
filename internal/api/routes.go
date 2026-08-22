package api

import "net/http"

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", makeHTTPHandleFunc(s.handleHealth))
	mux.HandleFunc("GET /hello", makeHTTPHandleFunc(s.handleMessage))
	mux.HandleFunc("GET /logs", makeHTTPHandleFunc(s.handleTokenLog))
	return mux
}
