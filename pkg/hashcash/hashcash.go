package hashcash

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"
)

const (
	zeroByte = 48

	// randLimit is the maximum number that can be generated.
	randLimit = 1000
)

// Errors.
var (
	ErrMaxIterationsExceeded = errors.New("maximum iterations exceeded")
)

// Hashcash is struct with fields of Hashcash.
type Hashcash struct {
	zeroBits int    // Number of leading zero bits in the hash
	counter  int    // Iteration counter
	rand     []byte // Random string
	header   []byte // Hashcash header without counter part
}

// New creates a new Hashcash struct.
func New(zeroBits int, resource, rand []byte) *Hashcash {
	buf := &bytes.Buffer{}

	// Version.
	buf.WriteByte(byte(FirstVersion))
	//Delimiter.
	buf.WriteByte(delimiter)
	// Bits.
	buf.Write([]byte(strconv.Itoa(zeroBits)))
	//Delimiter.
	buf.WriteByte(delimiter)
	// Date.
	buf.Write([]byte(time.Now().Format(dateFormat)))
	//Delimiter.
	buf.WriteByte(delimiter)
	// Resource.
	buf.Write(resource)
	// Double Delimiter.
	buf.Write([]byte{delimiter, delimiter})
	// Random.
	buf.Write(rand)
	// Delimiter.
	buf.WriteByte(delimiter)

	return &Hashcash{
		zeroBits: zeroBits,
		counter:  0,
		rand:     rand,
		header:   buf.Bytes(),
	}
}

// Generates a random encoded in base-64 format.
func GenerateRandom() []byte {
	// Generate a random number
	num, _ := rand.Int(rand.Reader, big.NewInt(randLimit))

	// Convert the random number to a byte slice
	randomBytes := []byte(num.String())

	// Encode the byte slice in base-64
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(randomBytes)))
	base64.StdEncoding.Encode(encoded, randomBytes)

	return encoded
}

// Header returns a byte slice representation of HashcashData
func (h Hashcash) Header() []byte {
	counterBytes := []byte(strconv.Itoa(h.counter))
	base64Counter := make([]byte, base64.StdEncoding.EncodedLen(len(counterBytes)))
	base64.StdEncoding.Encode(base64Counter, counterBytes)

	h.header = append(h.header, base64Counter...)

	return h.header
}

// GetCounter returns the current iteration counter.
func (h *Hashcash) GetCounter() int {
	return h.counter
}

// GetBits returns the number of leading zero bits in the hash.
func (h *Hashcash) GetBits() int {
	return h.zeroBits
}

// GetRand returns the random part.
func (h *Hashcash) GetRand() []byte {
	return h.rand
}

// IsValidSolution checks if the given hash has the required number of leading zeros.
func (h *Hashcash) isValid(hash []byte) bool {
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
func calculateHash(data []byte) []byte {
	hash := sha1.Sum(data)
	return []byte(fmt.Sprintf("%x", hash))
}

// Validate checks if the Header is valid.
func (h *Hashcash) Validate() bool {
	return h.isValid(calculateHash(h.Header()))
}

// Calculate generates a valid Hashcash based on the provided criteria.
func (h *Hashcash) Calculate(maxIterations int) error {
	// Calculate the hash until the number of leading zero bits is reached.
	for h.counter <= maxIterations {
		hash := calculateHash(h.Header())

		if h.isValid(hash) {
			return nil
		}

		h.counter++
	}

	return ErrMaxIterationsExceeded
}
