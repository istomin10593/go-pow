package hashcash

import (
	"bytes"
	"encoding/base64"
	"errors"
	"strconv"
)

// Parsing errors.
var (
	ErrInvalidDelimiter = errors.New("invalid delimiter")
	ErrInvalidVersion   = errors.New("invalid version")
	ErrInvalidBits      = errors.New("invalid bits")
	ErrInvalidRand      = errors.New("invalid rand")
	ErrInvalidCounter   = errors.New("invalid counter")
)

// The header line follows this format:
//
// Ver:Bits:Date:Resource:Rand:Counter
//
// It looks something like this:
//
// 1:20:1303030600:255.255.0.0:80::McMybZIhxKXu57jd:ckvi
//
// The header contains:
//
// Ver: Hashcash format version, 1 (which supersedes version 0).
// Bits: Number of "partial pre-image" (zero) bits.
// Date: The time that the message was sent, in the format YYMMDD[hhmm[ss]].
// Resource: Resource data string being transmitted an IP address.
// Rand: String of random characters, encoded in base-64 format.
// Counter: Binary counter, encoded in base-64 format.

// Delimiter
const delimiter byte = 58

// Version is a type for Hashcash format version.
type VersionProt byte

// Hashcash format version 1 with value 49 in byte representation.
const FirstVersion VersionProt = 49

// Date formats for Hashcash.
const dateFormat = "060102150405"

// Parse parses a Hashcash header into a Hashcash struct.
func (h *Hashcash) Parse(header []byte) error {
	// Split the header using the double delimiter.
	delimiterIndex := bytes.Index(header, []byte{delimiter, delimiter})
	if delimiterIndex == -1 {
		return ErrInvalidDelimiter
	}

	firstPart := header[:delimiterIndex]
	secondPart := header[delimiterIndex+2:]

	// Validate version.
	if VersionProt(firstPart[0]) != FirstVersion {
		return ErrInvalidVersion
	}

	// Check the first occurrence of the delimiter.
	if firstPart[1] != delimiter {
		return ErrInvalidDelimiter
	}

	// Validate bits.
	delimiterIndex = bytes.Index(firstPart[2:], []byte{delimiter})
	if delimiterIndex == -1 {
		return ErrInvalidDelimiter
	}

	bits, err := strconv.Atoi(string(firstPart[2 : 2+delimiterIndex]))
	if err != nil || bits < 1 {
		return ErrInvalidBits
	}

	// Split the second part using the single delimiter.
	delimiterIndex = bytes.Index(secondPart, []byte{delimiter})
	if delimiterIndex == -1 {
		return ErrInvalidDelimiter
	}

	rundBase64 := secondPart[:delimiterIndex]
	counterBase64 := secondPart[delimiterIndex+1:]

	// Decode random value from base64.
	_, err = base64.StdEncoding.DecodeString(string(rundBase64))
	if err != nil {
		return ErrInvalidRand
	}

	// Convert counter from base64 to an integer.
	counter, err := toCounter(counterBase64)
	if err != nil {
		return ErrInvalidCounter
	}

	// Update Hashcash struct fields.
	h.zeroBits = bits
	h.rand = rundBase64
	h.counter = counter
	h.header = header[:len(header)-len(counterBase64)]

	return nil
}

// toCounter parses a counter string in base-64 format into an integer.
func toCounter(counter []byte) (int, error) {
	// Decode the base64-encoded string.
	decodedBytes, err := base64.StdEncoding.DecodeString(string(counter))
	if err != nil {
		return 0, err
	}

	// Convert the decoded bytes to an integer.
	decodedInt, err := strconv.Atoi(string(decodedBytes))
	if err != nil {
		return 0, err
	}

	return decodedInt, nil
}
