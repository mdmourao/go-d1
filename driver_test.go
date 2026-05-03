package god1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDSNInvalid(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		wantErr error
	}{
		{
			name:    "empty DSN",
			dsn:     "",
			wantErr: ErrDriverMissingDSN,
		},
		{
			name:    "invalid DSN format",
			dsn:     "://invalid-dsn",
			wantErr: ErrDriverInvalidDSN,
		},
		{
			name:    "unsupported scheme",
			dsn:     "ftp://example.com",
			wantErr: ErrDriverInvalidDSNScheme,
		},
		{
			name:    "missing host",
			dsn:     "http://",
			wantErr: ErrDriverInvalidDSNHost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseDSN(tt.dsn)
			assert.Error(t, err)
		})
	}
}

func TestParseDSNValid(t *testing.T) {
	dsn := "https://example.com"
	u, err := parseDSN(dsn)
	assert.NoError(t, err)
	assert.Equal(t, "https", u.Scheme)
	assert.Equal(t, "example.com", u.Host)
}

func TestOpen(t *testing.T) {
	d := &d1Driver{}

	_, err := d.Open("invalid-dsn")
	assert.Error(t, err)

	_, err = d.Open("https://example.com")
	assert.NoError(t, err)
}

func TestOpenConnector(t *testing.T) {
	d := &d1Driver{}

	_, err := d.OpenConnector("invalid-dsn")
	assert.Error(t, err)

	connector, err := d.OpenConnector("https://example.com")
	assert.NoError(t, err)
	assert.NotNil(t, connector)
}
