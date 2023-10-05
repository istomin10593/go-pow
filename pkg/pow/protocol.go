package pow

import (
	"bytes"
	"errors"
	"net"
	"time"
)

var (
	// Handling errors.
	ErrReadConn  = errors.New("failed to read from connection")
	ErrWriteConn = errors.New("failed to write to connection")

	// Parsing errors.
	ErrInvalidMessage   = errors.New("invalid protocol message")
	ErrInvalidPhase     = errors.New("invalid phase")
	ErrInvalidDelimiter = errors.New("invalid delimiter")
)

// The PoW communication protocol Message follows this format:
//
// `Phase_Payload`
//
// Components:
// - `Phase` - Communication protocol phase.
// - 0 represents the initial phase where the server issues a challenge to the client.
// - 1 represents the validation phase where the server validates the client's solution.
//
// `_` - Delimiter separates the Phase and the Payload.
//
// - `Payload`: Contains various information, including challenge headers, provided tokens, and other messages.
//   - This section can include different types of data depending on the protocol phase.
//   - For example, during the initial phase, it may carry a challenge header.
//   - In the validation phase, it could contain a provided token for verification or requested information.
//   - Other types of messages or data may also be present, depending on the specific use case.

// Delimiter.
var delimiter byte = 95

// PhaseProt represents the PoW protocol Phase.
type PhaseProt byte

const (
	// InitPhase represents the PoW protocol used during the initialization Phase.
	InitPhase PhaseProt = 48

	// ValidPhase represents the PoW protocol used during the validation Phase.
	ValidPhase PhaseProt = 49
)

// Default timeout for reading.
const (
	defaultTimeout = time.Second * 5
)

// Protocol structure.
type Protocol struct {
	conn    net.Conn      // conn is a connection adhering to the standard connection interface.
	timeout time.Duration // timeout is the duration for reading steps.
	phase   PhaseProt     // phase represents the communication phase.
	payload []byte        // payload is the message payload.
}

// New returns a new Protocol instance.
func New(conn net.Conn, timeout time.Duration) *Protocol {
	if timeout == 0 {
		timeout = defaultTimeout
	}

	return &Protocol{
		conn:    conn,
		timeout: timeout,
		phase:   InitPhase,
		payload: make([]byte, 0),
	}
}

// String returns the protocol message as a string.
func (p *Protocol) Message() []byte {
	buf := &bytes.Buffer{}

	buf.WriteByte(byte(p.Phase()))
	buf.WriteByte(delimiter)
	buf.Write(p.Payload())

	return buf.Bytes()
}

// Phase returns the protocol Phase.
func (p *Protocol) Phase() PhaseProt {
	return p.phase
}

// SetPhase sets  the protocol Phase.
func (p *Protocol) SetPhase(phase PhaseProt) {
	p.phase = phase
}

// Payload returns the protocol Payload.
func (p *Protocol) Payload() []byte {
	return p.payload
}

// SetPayload sets  the protocol Payload.
func (p *Protocol) SetPayload(payload []byte) {
	p.payload = payload
}

// Read reads the protocol message from the connection.
func (p *Protocol) Read() error {
	// Set a timeout for reading.
	p.conn.SetReadDeadline(time.Now().Add(p.timeout))

	// Read the mesage.
	readBuf := make([]byte, 1024)
	n, err := p.conn.Read(readBuf)
	if err != nil {
		return ErrReadConn
	}

	readBuf = readBuf[:n]

	// Parse the message.
	// Massage must be at least 2 bytes long - Phase and Delimiter.
	if len(readBuf) < 2 {
		return ErrInvalidMessage
	}

	// Delimiter is mandatory.
	if readBuf[1] != delimiter {
		return ErrInvalidDelimiter
	}

	// Check the Phase.
	switch PhaseProt(readBuf[0]) {
	case InitPhase:
		// Set phase to Initial Phase.
		p.phase = InitPhase
		// Set payload if existed.
		if len(readBuf) > 2 {
			p.payload = readBuf[2:]
		}

	case ValidPhase:
		// Validation phase must have a payload.
		if len(readBuf) < 3 {
			return ErrInvalidMessage
		}

		// Set phase to Validation Phase and assign payload.
		p.phase = ValidPhase
		p.payload = readBuf[2:]
	default:
		// Invalid phase.
		return ErrInvalidPhase
	}

	return nil
}

// Write writes the protocol message to the connection.
func (p *Protocol) Write() error {
	if _, err := p.conn.Write(p.Message()); err != nil {
		return ErrWriteConn
	}

	return nil
}
