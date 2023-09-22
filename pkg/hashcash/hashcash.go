package hashcash

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"time"
)

const zeroByte = 48

// Errors for Hashcash.
var (
	ErrMaxIterationsExceeded = errors.New("maximum iterations exceeded")
)

// Hashcash is struct with fields of Hashcash.
type Hashcash struct {
	version  VersionProt // Version of Hashcash
	zeroBits int         // Number of leading zero bits in the hash
	date     string      // Timestamp
	resource string      // Resource identifier(host+port)
	rand     string      // Random value
	counter  int         // Iteration counter
}

// New creates a new Hashcash struct.
func New(zeroBits int, resource, rand string) *Hashcash {
	return &Hashcash{
		version:  FirstVersion,
		zeroBits: zeroBits,
		date:     time.Now().Format(dateFormat),
		resource: resource,
		rand:     rand,
		counter:  0,
	}
}

// String returns a string representation of HashcashData
func (h Hashcash) String() string {
	return fmt.Sprintf(
		"%d:%d:%s:%s::%s:%s",
		h.version,
		h.zeroBits,
		h.date,
		h.resource,
		h.rand,
		base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(h.counter))),
	)
}

// GetCounter returns the current iteration counter.
func (h *Hashcash) GetCounter() int {
	return h.counter
}

// GetResource returns the resource identifier.
func (h *Hashcash) GetResource() string {
	return h.resource
}

// IsValidSolution checks if the given hash has the required number of leading zeros.
func (h *Hashcash) isValid(hash string) bool {
	if h.zeroBits > len(hash) {
		return false
	}

	for _, bit := range hash[:h.zeroBits] {
		if bit != zeroByte {
			return false
		}
	}

	return true
}

// calculateHash calculates the SHA-1 hash of the given data.
func calculateHash(data string) string {
	hash := sha1.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// Calculate generates a valid Hashcash based on the provided criteria.
func (h *Hashcash) Calculate(maxIterations int) error {
	// Check if the header is valid.
	if maxIterations == 0 {
		if h.isValid(calculateHash(h.String())) {
			return nil
		} else {
			return ErrInvalidHeader
		}
	}

	// Calculate the hash until the number of leading zero bits is reached.
	for h.counter <= maxIterations {
		hash := calculateHash(h.String())

		if h.isValid(hash) {
			return nil
		}

		h.counter++
	}

	return ErrMaxIterationsExceeded
}
