package server

import (
	hash "go-pow/pkg/hashcash"
	prot "go-pow/pkg/proto"
	"go-pow/server/pkg/random"
	"net"
	"time"

	"go.uber.org/zap"
)

// handleClient handles a single connection.
func (s *Server) handle(conn net.Conn) error {
	defer conn.Close()

	// Read client's request.
	request, err := s.readMessage(conn)
	if err != nil {
		return err
	}

	// 	Create new message.
	var message *prot.Message

	message.Parse(request)

	var response string

	switch message.GetPhase() {
	case prot.InitPhase:
		response, err = s.initPhase(conn, message.GetHeader())
		if err != nil {
			return err
		}
	case prot.ValidPhase:
		response, err = s.validPhase(conn, message.GetHeader())
		if err != nil {
			return err
		}
	}

	message.SetPhase(prot.ValidPhase)
	message.SetHeader(response)

	// Send the response to the client.
	if err := s.writeMessage(conn, message.String()); err != nil {
		return err
	}

	return nil
}

// readMessage  reads a request from the client.
func (s *Server) readMessage(conn net.Conn) (string, error) {
	// Set a timeout for reading the client's request.
	conn.SetReadDeadline(time.Now().Add(s.cfg.Server.Timeout))

	// Read the mesage from the client.
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		s.log.Error("failed to read a message", zap.Error(err))

		return "", err
	}

	return string(buf[:n]), nil
}

// writeMessage writes a response to the client.
func (s *Server) writeMessage(conn net.Conn, response string) error {
	if _, err := conn.Write([]byte(response)); err != nil {
		s.log.Error("failed to write a message", zap.Error(err))

		return err
	}

	return nil
}

// initPhase handles a connection with init phase.
func (s *Server) initPhase(conn net.Conn, header string) (string, error) {
	// Read client's IP and port.
	clientIP, clientPort, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		s.log.Error("failed to get client's IP and port", zap.Error(err))

		return "", err
	}

	// Generate a random seed.
	rand := random.New()

	// Concatenate client IP and port.
	resource := clientIP + ":" + clientPort

	// Create a hashcash instance.
	hashcash := hash.New(s.cfg.Pow.ZeroBits, resource, rand)

	s.log.Debug("prepared a challenge header", zap.String("header", hashcash.String()))

	return hashcash.String(), nil
}

// validPhase handles a connection with valid phase.
func (s *Server) validPhase(conn net.Conn, header string) (string, error) {
	s.log.Debug("received a solution header", zap.String("header", header))

	// Check if the solution is valid.
	var hashcashResponse hash.Hashcash

	if err := hashcashResponse.Parse(header); err != nil {
		s.log.Error("failed to parse a solution header", zap.Error(err))

		return "", err
	}

	if err := hashcashResponse.Calculate(0); err != nil {
		s.log.Error("pow validation failed", zap.Error(err))

		return "", err
	}

	s.log.Info("pow validation successful", zap.Any("resource", hashcashResponse.GetResource()))

	// Send a random quote to the client.

	return s.book.GetRandQuote(), nil
}
