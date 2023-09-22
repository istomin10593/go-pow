package prot

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// The communication message follows this format:
//
// `1_1:20:1303030600:255.255.0.0:80::McMybZIhxKXu57jd:ckvi`
//
// `Phase_Hashcash header`
//
// Components:
// `1` - Phase
// - 0 represents the initial phase where the server issues a challenge to the client.
// - 1 represents the validation phase where the server validates the client's solution.
//
// `_` - Delimiter:
// - Separates the phase and the Hashcash header.
//
// `1:20:1303030600:255.255.0.0:80::McMybZIhxKXu57jd:ckvi` - Hashcash Header:
// - Contains information for solving the server's challenge and the client's validation request.

var (
	ErrInvalidMessage   = errors.New("invalid message")
	ErrInvalidPhase     = errors.New("invalid phase")
	ErrInvalidDelimiter = errors.New("invalid delimiter")
)

// PhaseProt is a type for communication protocol.
type PhaseProt int

const (
	// InitPhase represents the communication protocol used during the initialization phase.
	InitPhase PhaseProt = 0

	// ValidPhase represents the communication protocol used during the validation phase.
	ValidPhase PhaseProt = 1
)

// Communication message.
type Message struct {
	phase  PhaseProt
	header string
}

// String returns the message as a string.
func (m *Message) String() string {
	return fmt.Sprintf("%d_%s", m.phase, m.header)
}

// GetHeader returns the message header.
func (m *Message) GetPhase() PhaseProt {
	return m.phase
}

// GetHeader returns the message header.
func (m *Message) GetHeader() string {
	return m.header
}

// SetPhase sets the message phase.
func (m *Message) SetPhase(phase PhaseProt) {
	m.phase = phase
}

// SetHeader sets the message header.
func (m *Message) SetHeader(header string) {
	m.header = header
}

// Parse parses the message according to the protocol.
func (m *Message) Parse(msg string) error {
	// Split the message by the delimiter '_'
	parts := strings.Split(msg, "_")
	if len(parts) == 0 {
		return ErrInvalidDelimiter
	}

	// Parse and populate the message.
	phase, err := strconv.Atoi(parts[0])
	if err != nil {
		return ErrInvalidPhase
	}

	switch PhaseProt(phase) {
	case InitPhase:
		if len(parts) != 1 {
			return ErrInvalidMessage
		}

		m.phase = PhaseProt(phase)
	case ValidPhase:
		if len(parts) != 2 {
			return ErrInvalidMessage
		}

		m.phase = PhaseProt(phase)
		m.header = parts[1]
	default:
		return ErrInvalidPhase
	}

	return nil
}
