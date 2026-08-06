package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/chishkin/intask/notifier/internal/infrastructure/config"
)

type Server struct {
	cfg *config.Config
	log *slog.Logger
	srv *http.Server
}

// New creates a Server bound to the address in cfg.Server.HTTP.Addr.
// It returns an error if required configuration is missing or invalid.
//
// The handler is invoked for every incoming request; wrap it with middleware
// (logging, recovery, timeouts) before passing to New.
func New(cfg *config.Config, log *slog.Logger, handler http.Handler) *Server {
	return &Server{
		cfg: cfg,
		log: log,
		srv: &http.Server{
			Addr:         cfg.Server.HTTP.Addr,
			Handler:      handler,
			ReadTimeout:  cfg.Server.Conns.ReadTimeout,
			IdleTimeout:  cfg.Server.Conns.IdleTimeout,
			WriteTimeout: cfg.Server.Conns.WriteTimeout,
		},
	}
}

// Start begins accepting connections. It blocks until Shutdown is called
// or a fatal error occurs. http.ErrServerClosed is treated as a clean exit
// and not returned as an error.
//
// Start is typically paired with Shutdown via signal handling:
//
//	go func() { _ = srv.Start() }()
//	<-stop
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//	_ = srv.Shutdown(ctx)
func (s *Server) Start() error {
	s.log.Info("server is running...",
		slog.String("addr", s.cfg.Server.HTTP.Addr),
	)

	if s.cfg.Server.HTTP.TLS.Enable {
		s.log.Info("tls enabled",
			slog.String("server_cert", s.cfg.Server.HTTP.TLS.ServerCertPath),
			slog.String("server_key", s.cfg.Server.HTTP.TLS.ServerKeyPath),
		)

		return s.srv.ListenAndServeTLS(
			s.cfg.Server.HTTP.TLS.ServerCertPath,
			s.cfg.Server.HTTP.TLS.ServerKeyPath,
		)
	}

	return s.srv.ListenAndServe()
}

// Shutdown gracefully stops the server: no new connections are accepted,
// in-flight requests are served until ctx expires, then the listener closes.
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("server shutdown")
	return s.srv.Shutdown(ctx)
}
