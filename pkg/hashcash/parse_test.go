package hashcash

import (
	"encoding/base64"
	"strconv"
	"testing"
	"time"

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
			wantErr: ErrInvalidHeader,
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
			name:    "Invalid Date",
			header:  "1:20:1303:255.255.0.0:80::NTQ2:MA==",
			wantErr: ErrInvalidDate,
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
			err := h.Parse(tt.header)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.Equal(t, tt.wantHashcash, h)
			}
		})
	}
}

func TestToDate(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		want    time.Time
		wantErr error
	}{
		{
			name:    "Valid date - YYMMDDhhmmss",
			date:    "060102150405",
			want:    time.Date(2006, 01, 02, 15, 04, 05, 00, time.UTC),
			wantErr: nil,
		},
		{
			name:    "Valid date - YYMMDDhhmm",
			date:    "0601021504",
			want:    time.Date(2006, 01, 02, 15, 04, 00, 00, time.UTC),
			wantErr: nil,
		},
		{
			name:    "Valid date - YYMMDD",
			date:    "060102",
			want:    time.Date(2006, 01, 02, 00, 00, 00, 00, time.UTC),
			wantErr: nil,
		},
		{
			name:    "Invalid date - YYMM",
			date:    "0601",
			want:    time.Time{},
			wantErr: ErrInvalidDate,
		},
		{
			name:    "Invalid date - invalid mounth",
			date:    "061302150405",
			want:    time.Time{},
			wantErr: ErrInvalidDate,
		},
		{
			name:    "Invalid date - invalid day",
			date:    "060132150405",
			want:    time.Time{},
			wantErr: ErrInvalidDate,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toDate(tt.date)
			assert.Equal(t, tt.want, got)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
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
			got, err := toCounter(tt.counter)
			assert.Equal(t, tt.want, got)
			if tt.wantErr != nil {
				assert.IsType(t, tt.wantErr, err)
			}
		})
	}
}
