package server

import (
	"context"
	"net/http"

	"github.com/andrearcaina/hyperion/internal/db"
	"github.com/andrearcaina/hyperion/internal/logger"
	http2 "github.com/andrearcaina/hyperion/internal/transport/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	srv    *http.Server
	db     *db.DB
	logger *logger.Logger
}

type ServerConfig struct {
	Port   string
	DBPath string
	Logger *logger.Logger
}

func NewServer(cfg *ServerConfig) (*Server, error) {
	db, err := db.New(cfg.DBPath)
	if err != nil {
		return nil, err
	}

	router := chi.NewRouter()

	router.Use(
		cfg.Logger.RequestLogger,
		middleware.Recoverer,
	)

	handler := http2.NewHandler(db, cfg.Logger)
	router.Mount("/hypr", handler.ServeRoutes())

	return &Server{
		srv: &http.Server{
			Addr:    cfg.Port,
			Handler: router,
		},
		db:     db,
		logger: cfg.Logger,
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

	if err := s.db.Close(); err != nil {
		return err
	}

	s.logger.Info(context.Background(), "Server closed gracefully")
	return nil
}
