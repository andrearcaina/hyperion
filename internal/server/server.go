package server

import (
	"context"
	"net/http"

	"github.com/andrearcaina/hyperion/internal/logger"
	http2 "github.com/andrearcaina/hyperion/internal/transport/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	srv    *http.Server
	logger *logger.Logger
}

func NewServer(port string, logger *logger.Logger, handler *http2.Handler) (*Server, error) {
	router := chi.NewRouter()

	router.Use(
		logger.RequestLogger,
		middleware.Recoverer,
	)

	router.Mount("/hypr", handler.ServeRoutes())

	return &Server{
		srv: &http.Server{
			Addr:    port,
			Handler: router,
		},
		logger: logger,
	}, nil
}

func (s *Server) Run() error {
	s.logger.Info(context.Background(), "Starting server on", "port", s.srv.Addr)

	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Server) Close(ctx context.Context) error {
	s.logger.Info(context.Background(), "Shutting down server gracefully...")

	err := s.srv.Shutdown(ctx)
	if err != nil {
		return err
	}

	s.logger.Info(context.Background(), "Server closed gracefully")
	return nil
}
