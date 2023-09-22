package hashcash

import (
	"encoding/base64"
	"errors"
	"regexp"
	"strconv"
	"time"
)

// The header line looks something like this:
//
// 1:20:1303030600:255.255.0.0:80::McMybZIhxKXu57jd:ckvi
//
// The header contains:
//
// ver: Hashcash format version, 1 (which supersedes version 0).
// bits: Number of "partial pre-image" (zero) bits in the hashed code.
// date: The time that the message was sent, in the format YYMMDD[hhmm[ss]].
// resource: Resource data string being transmitted an IP address.
// rand: String of random characters, encoded in base-64 format.
// counter: Binary counter, encoded in base-64 format.

// Version is a type for Hashcash format version.
type VersionProt int

const (
	// Hashcash format version 1.
	FirstVersion VersionProt = 1
)

var (
	// Date formats for Hashcash.
	dateFormat = "060102150405"
)

var (
	// Valid data formats.
	dateFormats = []string{
		"060102",       // YYMMDD
		"0601021504",   // YYMMDDhhmm
		"060102150405", // YYMMDDhhmmss
	}
)

// Regular expression for parsing the header line.
var (
	regexpHeader = regexp.MustCompile(`^(\d):(\d+):(\d+):(\d+\.\d+\.\d+\.\d+:\d+)::([A-Za-z0-9+/]+={0,2}):([A-Za-z0-9+/]+={0,2})$`)
)

// Errors.
var (
	ErrInvalidHeader  = errors.New("invalid header")
	ErrInvalidVersion = errors.New("invalid version")
	ErrInvalidBits    = errors.New("invalid bits")
	ErrInvalidDate    = errors.New("invalid date")
	ErrInvalidCounter = errors.New("invalid counter")
)

// Parse parses a Hashcash header into a Hashcash struct.
func (h *Hashcash) Parse(header string) error {
	matches := regexpHeader.FindStringSubmatch(header)

	if len(matches) != 7 {
		return ErrInvalidHeader
	}

	ver, err := strconv.Atoi(matches[1])
	if err != nil || VersionProt(ver) != FirstVersion {
		return ErrInvalidVersion
	}

	bits, err := strconv.Atoi(matches[2])
	if err != nil || bits < 1 {
		return ErrInvalidBits
	}

	date, err := toDate(matches[3])
	if err != nil {
		return err
	}

	counter, err := toCounter(matches[6])
	if err != nil {
		return ErrInvalidCounter
	}

	h.version = VersionProt(ver)
	h.zeroBits = bits
	h.date = date.Format(dateFormat)
	h.resource = matches[4]
	h.rand = matches[5]
	h.counter = counter

	return nil
}

// toDate parses a date string into a time.Time.
func toDate(date string) (time.Time, error) {
	for _, format := range dateFormats {
		date, err := time.Parse(format, date)
		if err == nil {
			return date, nil
		}
	}

	return time.Time{}, ErrInvalidDate
}

// toCounter parses a counter string in base-64 format into an integer.
func toCounter(counter string) (int, error) {
	// Decode the base64-encoded string
	decodedBytes, err := base64.StdEncoding.DecodeString(counter)
	if err != nil {
		return 0, err
	}

	// Convert the decoded bytes to an integer
	decodedInt, err := strconv.Atoi(string(decodedBytes))
	if err != nil {
		return 0, err
	}

	return decodedInt, nil
}
