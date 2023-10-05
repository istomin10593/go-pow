package hashcash

import (
	"encoding/base64"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name         string
		header       string
		wantHashcash *Hashcash
		wantErr      error
	}{
		{
			name:         "Valid header",
			header:       "1:20:1303030600:255.255.0.0:80::NTQ2:MA==",
			wantHashcash: getHashcash(),
		},
		{
			name:    "Invalid header",
			header:  "1:20:1303030600:255.255.0.0:80::MA==",
			wantErr: ErrInvalidDelimiter,
		},
		{
			name:    "Invalid version",
			header:  "0:20:1303030600:255.255.0.0:80::NTQ2:MA==",
			wantErr: ErrInvalidVersion,
		},
		{
			name:    "Invalid Bits",
			header:  "1:0:1303030600:255.255.0.0:80::NTQ2:MA==",
			wantErr: ErrInvalidBits,
		},
		{
			name:    "Invalid header",
			header:  "1:20:1303030600:255.255.0.0:80::_:MA==",
			wantErr: ErrInvalidRand,
		},
		{
			name:    "Invalid counter",
			header:  "1:20:1303030600:255.255.0.0:80::NTQ2:NTQ2==",
			wantErr: ErrInvalidCounter,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Hashcash{}
			err := h.Parse([]byte(tt.header))
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.Equal(t, tt.wantHashcash, h)
			}
		})
	}
}

func TestToCounter(t *testing.T) {
	tests := []struct {
		name    string
		counter string
		want    int
		wantErr error
	}{
		{
			name:    "Valid counter",
			counter: "MTIzNA==",
			want:    1234,
			wantErr: nil,
		},
		{
			name:    "Invalid base64 encoding",
			counter: "NTQ2==",
			want:    0,
			wantErr: base64.CorruptInputError(0),
		},
		{
			name:    "Invalid integer format",
			counter: "YWJjZA==",
			want:    0,
			wantErr: &strconv.NumError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toCounter([]byte(tt.counter))
			assert.Equal(t, tt.want, got)
			if tt.wantErr != nil {
				assert.IsType(t, tt.wantErr, err)
			}
		})
	}
}
