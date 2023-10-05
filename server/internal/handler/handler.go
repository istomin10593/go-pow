package handler

import (
	"errors"
	"fmt"

	"go-pow/pkg/hashcash"
	hash "go-pow/pkg/hashcash"
	"go-pow/pkg/pow"
	"go-pow/server/pkg/book"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	ErrInvalidSolutionHeader = errors.New("invalid solution header")
	ErrUnreconizedRandom     = errors.New("rand is not in the cache")
)

// Handler.
type Handler struct {
	zeroBits int
	timeout  time.Duration
	mu       sync.RWMutex
	log      *zap.Logger
	book     *book.Book
	cache    Cache
}

// Cache interface.
type Cache interface {
	Add([]byte) error
	Get([]byte) (bool, error)
	Delete([]byte) error
}

// New creates a new Handler instance.
func New(
	zeroBits int,
	timeout time.Duration,
	log *zap.Logger,
	book *book.Book,
	cache Cache,
) *Handler {
	return &Handler{
		zeroBits: zeroBits,
		timeout:  timeout,
		mu:       sync.RWMutex{},
		log:      log,
		book:     book,
		cache:    cache,
	}
}

// handleClient handles a single connection.
func (h *Handler) Handle(conn net.Conn) error {
	defer conn.Close()

	// New pow protocol instance.
	prot := pow.New(conn, h.timeout)

	// Read client's request.
	if err := prot.Read(); err != nil {
		h.log.Error("failed to read request", zap.Error(err))

		return fmt.Errorf("failed to read request: %w", err)
	}

	// 	Create new payload.
	var payload []byte
	var err error

	switch prot.Phase() {
	case pow.InitPhase:
		// Get payload with a challenge header.
		payload, err = h.init([]byte(conn.RemoteAddr().String()))
		if err != nil {
			h.log.Error("failed to init a challenge  header", zap.Error(err))

			return fmt.Errorf("failed to init a challenge  header: %w", err)
		}
	case pow.ValidPhase:
		// Validate the solution header
		// and get payload with valuable information.
		payload, err = h.valid(prot.Payload())
		if err != nil {
			h.log.Error("failed to validate soution header", zap.Error(err))

			return fmt.Errorf("failed to validate solution header: %w", err)
		}
	}

	// Set payload to the response.
	prot.SetPayload(payload)

	// Send the response to the client.
	if err := prot.Write(); err != nil {
		h.log.Error("failed to write response", zap.Error(err))

		return fmt.Errorf("failed to write response: %w", err)
	}

	return nil
}

// init handles a connection with init phase.
func (h *Handler) init(address []byte) ([]byte, error) {
	// Generate a random seed.
	rand := hashcash.GenerateRandom()

	// Create a hashcash instance.
	hashcash := hash.New(h.zeroBits, address, rand)

	// Add rand to the cache.
	if err := h.cache.Add(rand); err != nil {
		h.log.Error("failed to add rand to the cache", zap.Error(err))

		return nil, fmt.Errorf("failed to add rand to the cache: %w", err)
	}

	// Prepare a challenge header.

	h.log.Debug("prepared a challenge header", zap.String("header", string(hashcash.Header())))

	return hashcash.Header(), nil
}

// validPhase handles a connection with valid phase.
func (h *Handler) valid(header []byte) ([]byte, error) {
	h.log.Debug("received a solution header", zap.ByteString("header", header))

	// Check if the solution is valid.
	var hashcashResponse hash.Hashcash

	// Parse the solution header.
	if err := hashcashResponse.Parse(header); err != nil {
		h.log.Error("failed to parse a solution header", zap.Error(err))

		return nil, fmt.Errorf("failed to parse a solution header: %w", err)
	}

	// Check if the 'rand' value is in the cache.
	// If the 'rand' value is not found in the cache, it could indicate that either the challenge has expired
	// or the server did not provide the required challenge header for that 'rand' value.
	h.log.Debug("checking if rand is in the cache", zap.String("rand", string(hashcashResponse.GetRand())))

	ok, err := h.cache.Get(hashcashResponse.GetRand())
	if err != nil {
		h.log.Error("error checking cache", zap.Error(err))

		return nil, fmt.Errorf("error checking cache: %w", err)
	}

	if !ok {
		h.log.Error("rand is not in the cache")

		return nil, ErrUnreconizedRandom
	}

	// Validate the solution header by performing a single hashcash calculation.
	if !hashcashResponse.Validate() {
		h.log.Error("pow validation failed")

		return nil, ErrInvalidSolutionHeader
	}

	h.log.Info("pow validation successful", zap.String("header", string(hashcashResponse.Header())))

	// Send a random quote to the client.
	return h.book.GetRandQuote(), nil
}
