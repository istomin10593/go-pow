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

// Handler interface.
type Handler interface {
	Handle(net.Conn) error
}

// Server.
type Server struct {
	cfg  *config.Config
	book *book.Book
	hand Handler
}

// Creates a new server.
func New(cfg *config.Config, book *book.Book, hand Handler) *Server {
	return &Server{
		cfg:  cfg,
		book: book,
		hand: hand,
	}
}

// Run starts the server.
func (s *Server) Run(ctx context.Context, log *zap.Logger) error {
	ln, err := net.Listen("tcp", s.cfg.Server.Host+s.cfg.Server.Port)
	if err != nil {
		log.Error("failed to start server", zap.Error(err))

		return fmt.Errorf("failed to start server: %w", err)
	}

	log.Info("listening", zap.String("port", s.cfg.Server.Port))

	errCh := make(chan error)

	go func() {
		defer close(errCh)
		<-ctx.Done()

		log.Info("server is shutting down...")
		errCh <- ln.Close()
	}()

	for {
		conn, err := ln.Accept()

		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Info("listener closed intentionally")

				break
			}
			log.Error("failed to accept connection", zap.Error(err))

			continue
		}

		log.Debug("connection accepted", zap.String("client address", conn.RemoteAddr().String()))

		go func() {
			if err := s.hand.Handle(conn); err != nil {
				log.Info("handle connection failed", zap.String("client address", conn.RemoteAddr().String()))
			} else {
				log.Info("handle connection succeeded", zap.String("client address", conn.RemoteAddr().String()))
			}
		}()
	}

	select {
	case err := <-errCh:
		log.Error("failed to shutdown", zap.Error(err))

		return fmt.Errorf("failed to shutdown: %w", err)
	default:
		return nil
	}
}
