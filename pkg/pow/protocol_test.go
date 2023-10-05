package pow

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtocol_Read(t *testing.T) {
	test := []struct {
		name            string
		input           string
		expectedPhase   PhaseProt
		expectedPayload []byte
		expectedErr     error
	}{
		{
			name:          "Valid Message - Init Phase",
			input:         "0_",
			expectedPhase: InitPhase,
			expectedErr:   nil,
		},
		{
			name:            "Valid Message - Valid Phase",
			input:           "1_abc",
			expectedPhase:   ValidPhase,
			expectedPayload: []byte("abc"),
			expectedErr:     nil,
		},
		{
			name:          "Invalid Message",
			input:         "0",
			expectedPhase: InitPhase,
			expectedErr:   ErrInvalidMessage,
		},
		{
			name:          "Invalid Message - ommited payload for Valid Phase",
			input:         "1_",
			expectedPhase: InitPhase,
			expectedErr:   ErrInvalidMessage,
		},
		{
			name:          "Invalid Message (Missing Delimiter)",
			input:         "0abc",
			expectedPhase: InitPhase,
			expectedErr:   ErrInvalidDelimiter,
		},
		{
			name:          "Invalid Message (Invalid Phase)",
			input:         "2_abc",
			expectedPhase: InitPhase,
			expectedErr:   ErrInvalidPhase,
		},
	}

	for _, tc := range test {
		t.Run(tc.name, func(t *testing.T) {
			ln, err := net.Listen("tcp", "localhost:8081")
			assert.NoError(t, err)
			defer ln.Close()

			// Start a goroutine for the  client imitation.
			go func() {
				conn, err := net.Dial("tcp", ln.Addr().String())
				assert.NoError(t, err)
				defer conn.Close()

				// Simulate a valid message.
				conn.Write([]byte(tc.input))
			}()

			// Start a temporary test server to handle the connection.
			conn, err := ln.Accept()
			assert.NoError(t, err)
			defer conn.Close()

			// Create a new Protocol instance with a real connection.

			p := New(conn, defaultTimeout)

			// Test the Read method.
			err = p.Read()
			assert.Equal(t, tc.expectedPhase, p.phase)
			assert.Equal(t, tc.expectedPayload, p.payload)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}
