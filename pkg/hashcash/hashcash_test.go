package hashcash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// getHashcash generates a valid Hashcash struct.
func getHashcash() *Hashcash {
	return &Hashcash{
		zeroBits: 20,
		counter:  0,
		rand:     []byte("NTQ2"),
		header:   []byte("1:20:1303030600:255.255.0.0:80::NTQ2:"),
	}
}

func TestCalculate(t *testing.T) {
	tests := []struct {
		name          string
		zeroBits      int
		maxIterations int
		wantErr       error
	}{
		{
			name:          "Valid solution",
			zeroBits:      3,
			maxIterations: 10000,
		},
		{
			name:          "Insufficient iterations",
			zeroBits:      5,
			maxIterations: 10,
			wantErr:       ErrMaxIterationsExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := getHashcash()
			h.zeroBits = tt.zeroBits

			err := h.Calculate(tt.maxIterations)

			if tt.wantErr == nil {
				assert.Nil(t, err)
				assert.True(t, h.Validate())
			} else {
				assert.Equal(t, tt.wantErr, err)
			}
		})
	}
}

func TestIsValidSolution(t *testing.T) {
	tests := []struct {
		name     string
		zeroBits int
		hash     string
		want     bool
	}{
		{
			name:     "Valid solution",
			zeroBits: 3,
			hash:     "000abcdef",
			want:     true,
		},
		{
			name:     "Invalid solution (insufficient leading zeros)",
			zeroBits: 4,
			hash:     "00abcdef",
			want:     false,
		},
		{
			name:     "Invalid solution (ZeroBits > hash length)",
			zeroBits: 6,
			hash:     "00abc",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := getHashcash()
			h.zeroBits = tt.zeroBits

			got := h.isValid([]byte(tt.hash))
			assert.Equal(t, tt.want, got)
		})
	}
}
