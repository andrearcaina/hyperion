package server

import (
	"context"
	"log"
	"net/http"

	"github.com/andrearcaina/hyperion/internal/db"
	http2 "github.com/andrearcaina/hyperion/internal/transport/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	srv *http.Server
	db  *db.DB
}

type ServerConfig struct {
	Port   string
	DBPath string
}

func NewServer(cfg *ServerConfig) (*Server, error) {
	db, err := db.New(cfg.DBPath)
	if err != nil {
		return nil, err
	}

	router := chi.NewRouter()

	router.Use(
		middleware.Logger,
		middleware.Recoverer,
	)

	handler := http2.NewHandler(db)
	router.Mount("/hypr", handler.ServeRoutes())

	return &Server{
		srv: &http.Server{
			Addr:    cfg.Port,
			Handler: router,
		},
		db: db,
	}, nil
}

func (s *Server) Run() error {
	log.Printf("Starting server on %s", s.srv.Addr)
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Server) Close(ctx context.Context) error {
	log.Printf("Shutting down server gracefully...")

	err := s.srv.Shutdown(ctx)
	if err != nil {
		return err
	}

	if err := s.db.Close(); err != nil {
		return err
	}

	log.Printf("Server closed gracefully")
	return nil
}
