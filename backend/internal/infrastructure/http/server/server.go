package server

import (
	"context"
	"net/http"

	"github.com/chishkin-afk/intask/backend/internal/infrastructure/config"
)

// Server represents an HTTP server instance.
//
// It manages the lifecycle of the underlying net/http server,
// including starting the server and graceful shutdown.
type Server struct {
	cfg *config.Config
	srv *http.Server
}

// New creates a new HTTP server with the provided configuration and request handler.
//
// The server configuration is used to initialize address, connection timeouts,
// and TLS settings. The provided handler will be used as the root HTTP handler.
func New(cfg *config.Config, handler http.Handler) *Server {
	return &Server{
		cfg: cfg,
		srv: &http.Server{
			Addr:         cfg.Server.HTTP.Addr,
			Handler:      handler,
			ReadTimeout:  cfg.Server.Conns.ReadTimeout,
			WriteTimeout: cfg.Server.Conns.WriteTimeout,
			IdleTimeout:  cfg.Server.Conns.IdleTimeout,
		},
	}
}

// Start starts the HTTP server.
//
// If TLS is enabled in the configuration, the server starts using HTTPS
// with the configured certificate and private key. Otherwise, it starts
// as a regular HTTP server.
//
// Start blocks until the server stops or an error occurs.
func (s *Server) Start() error {
	if s.cfg.Server.HTTP.TLS.Enable {
		return s.srv.ListenAndServeTLS(
			s.cfg.Server.HTTP.TLS.ServerCertPath,
			s.cfg.Server.HTTP.TLS.ServerKeyPath,
		)
	}

	return s.srv.ListenAndServe()
}

// Shutdown gracefully stops the HTTP server.
//
// It waits for active connections to finish before shutting down.
// The provided context controls the maximum time allowed for shutdown.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
