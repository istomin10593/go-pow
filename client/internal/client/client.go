package client

import (
	"context"

	"go-pow/client/pkg/config"
	hash "go-pow/pkg/hashcash"

	"net"
	"time"

	"go.uber.org/zap"
)

// Client.
type Client struct {
	cfg *config.Config
	log *zap.Logger
}

// Creates a new Client.
func New(log *zap.Logger, cfg *config.Config) *Client {
	return &Client{
		cfg: cfg,
		log: log,
	}
}

// Run starts the Client.
func (c *Client) Run(ctx context.Context) {
	done := make(chan struct{})
	defer close(done)

	for i := 0; i < c.cfg.Client.Number; i++ {
		go c.runConnection(ctx, done, i)

		time.Sleep(c.cfg.Client.Timeout)
	}

	for i := 0; i < c.cfg.Client.Number; i++ {
		<-done
	}
}

// runConnection handles a single connection.
func (c *Client) runConnection(ctx context.Context, done chan struct{}, ID int) {
	for {
		select {
		case <-ctx.Done():
			done <- struct{}{}

			return
		default:
			if err := handleConnection(c.log, c.cfg, done, ID); err != nil {
				c.log.Info("pow completed with error", zap.Int("clientID", ID))
			} else {
				c.log.Info("pow completed successfully", zap.Int("clientID", ID))
			}

			return
		}
	}
}

// handleConnection connects to the server and handles the pow.
func handleConnection(
	log *zap.Logger,
	cfg *config.Config,
	done chan struct{},
	ID int,
) error {
	defer func() {
		done <- struct{}{}
	}()

	log = log.With(zap.Int("clientID", ID))

	// Connect to the server.
	conn, err := net.Dial("tcp", cfg.Server.Host+cfg.Server.Port)
	if err != nil {
		log.Error("failed to connect to server", zap.Error(err))

		return err
	}
	defer conn.Close()

	log.Info("connected to", zap.String("address", cfg.Server.Host+cfg.Server.Port))

	// Set a timeout for reading the server responses.
	conn.SetReadDeadline(time.Now().Add(cfg.Server.Timeout))

	// Read the challenge header from the server.
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Error("failed to read a challenge response", zap.Error(err))

		return err
	}

	// Solve pow.
	headerPOW := string(buf[:n])

	log.Debug("got a challenge header", zap.String("headerPOW", headerPOW))

	var hashcash hash.Hashcash
	if err := hashcash.Parse(headerPOW); err != nil {
		log.Error("failed to parse the challenge header", zap.Error(err))

		return err
	}

	if err := hashcash.Calculate(cfg.Pow.MaxIterations); err != nil {
		log.Error("failed to solve the pow", zap.Error(err))

		return err
	}

	log.Debug("calculate the solution header", zap.String("solution", hashcash.String()))

	// Send the solution header to the server.
	if _, err := conn.Write([]byte(hashcash.String())); err != nil {
		log.Error("failed to send the solution header", zap.Error(err))

		return err
	}

	// Set a timeout for reading the server responses.
	conn.SetReadDeadline(time.Now().Add(cfg.Server.Timeout))

	// Read the response from the server.
	buf = make([]byte, 1024)
	n, err = conn.Read(buf)
	if err != nil {
		log.Error("failed to read a challenge response", zap.Error(err))

		return err
	}

	log.Info("received response from server", zap.String("response", string(buf[:n])))

	return nil
}
