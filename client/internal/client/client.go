package client

import (
	"context"

	"go-pow/client/pkg/config"
	hash "go-pow/pkg/hashcash"
	"go-pow/pkg/pow"

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
func (c *Client) runConnection(ctx context.Context, done chan struct{}, id int) {
	for {
		select {
		case <-ctx.Done():
			done <- struct{}{}

			return
		default:
			if err := c.handleConnection(id, done); err != nil {
				c.log.Info("pow completed with error", zap.Int("clientID", id))
			} else {
				c.log.Info("pow completed successfully", zap.Int("clientID", id))
			}

			return
		}
	}
}

// handleConnection connects to the server and handles the pow.
func (c *Client) handleConnection(
	id int,
	done chan struct{},
) error {
	defer func() {
		done <- struct{}{}
	}()

	log := c.log.With(zap.Int("clientID", id))

	connReq, err := net.Dial("tcp", c.cfg.Server.Host+c.cfg.Server.Port)
	if err != nil {
		c.log.Error("failed to connect to server", zap.Error(err))

		return err
	}
	defer connReq.Close()

	// New pow protocol instance for requesting a challenge header.
	prot := pow.New(connReq, c.cfg.Server.Timeout)

	// Request a challenge header by sending the message with init phase.
	if err = prot.Write(); err != nil {
		log.Error("failed to request a challenge header", zap.Error(err))

		return err
	}

	log.Debug("requested a challenge header")

	// Read a challenge header.
	if err = prot.Read(); err != nil {
		log.Error("failed to read a challenge header", zap.Error(err))

		return err
	}

	log.Debug("got a challenge header", zap.String("header", string(prot.Payload())))

	var hashcash hash.Hashcash
	if err = hashcash.Parse(prot.Payload()); err != nil {
		log.Error("failed to parse a challenge header", zap.Error(err))

		return err
	}

	if err = hashcash.Calculate(c.cfg.Pow.MaxIterations); err != nil {
		log.Error("failed to solve the pow", zap.Error(err))

		return err
	}

	log.Debug("calculate a solution header", zap.String("solution", string(hashcash.Header())))

	connResp, err := net.Dial("tcp", c.cfg.Server.Host+c.cfg.Server.Port)
	if err != nil {
		c.log.Error("failed to connect to server", zap.Error(err))

		return err
	}
	defer connResp.Close()

	// New pow protocol instance to respond with a solution header.
	prot = pow.New(connResp, c.cfg.Server.Timeout)

	// Populate the message with a solution header.
	prot.SetValidPhase()
	prot.SetPayload(hashcash.Header())

	// Send a solution header.
	if err = prot.Write(); err != nil {
		log.Error("failed to send a solution header", zap.Error(err))

		return err
	}

	// Read a challenge header.
	if err = prot.Read(); err != nil {
		log.Error("failed to read a wisdom quote", zap.Error(err))

		return err
	}

	log.Info("received response from server", zap.String("response", string(prot.Payload())))

	return nil
}
