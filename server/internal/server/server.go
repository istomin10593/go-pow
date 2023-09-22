package server

import (
	"context"
	"errors"
	"fmt"

	"go-pow/server/pkg/book"
	"go-pow/server/pkg/config"
	"net"

	"go.uber.org/zap"
)

// Server.
type Server struct {
	cfg  *config.Config
	log  *zap.Logger
	book *book.Book
}

// Creates a new server.
func New(log *zap.Logger, cfg *config.Config, book *book.Book) *Server {
	return &Server{
		cfg:  cfg,
		log:  log,
		book: book,
	}
}

// Run starts the server.
func (s *Server) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.cfg.Server.Host+s.cfg.Server.Port)
	if err != nil {
		s.log.Error("failed to start server", zap.Error(err))

		return fmt.Errorf("failed to start server: %w", err)
	}

	s.log.Info("listening", zap.String("port", s.cfg.Server.Port))

	errCh := make(chan error)

	go func() {
		defer close(errCh)
		<-ctx.Done()

		s.log.Info("server.Run: shutting down")
		errCh <- ln.Close()
	}()

	for {
		conn, err := ln.Accept()

		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				s.log.Info("listener closed intentionally")

				break
			}
			s.log.Error("failed to accept connection", zap.Error(err))

			continue
		}

		s.log.Debug("connection accepted", zap.String("client address", conn.RemoteAddr().String()))

		go func() {
			if err := s.handle(conn); err != nil {
				s.log.Info("handleConnection failed", zap.String("client address", conn.RemoteAddr().String()))
			} else {
				s.log.Info("handleConnection succeeded", zap.String("client address", conn.RemoteAddr().String()))
			}
		}()
	}

	select {
	case err := <-errCh:
		s.log.Error("failed to shutdown", zap.Error(err))

		return fmt.Errorf("failed to shutdown: %w", err)
	default:
		return nil
	}
}
