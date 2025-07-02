package export

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
)

func Test_defaultSshBannerReader_readSSHBanner(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		setupServer func(address string) func() // sets up a mock server for testing and returns a teardown function
		wantBanner  string
		wantErr     string
	}{
		{
			name:    "successful banner retrieval",
			address: "127.0.0.1:2222",
			setupServer: func(address string) func() {
				ln, err := net.Listen("tcp", address)
				require.NoError(t, err) // Ensure mock server starts successfully
				go func() {
					conn, _ := ln.Accept()
					defer conn.Close()
					_, _ = conn.Write([]byte("Welcome to the SSH server\n"))
				}()
				return func() { ln.Close() }
			},
			wantBanner: "Welcome to the SSH server\n",
			wantErr:    "",
		},
		{
			name:        "connection failure",
			address:     "127.0.0.1:9999",
			setupServer: func(address string) func() { return func() {} }, // No server setup
			wantBanner:  "",
			wantErr:     "connection failed while getting ssh-banner",
		},
		{
			name:    "read deadline failure",
			address: "127.0.0.1:2223",
			setupServer: func(address string) func() {
				ln, err := net.Listen("tcp", address)
				require.NoError(t, err)
				go func() {
					_, _ = ln.Accept()
					// Intentionally do not write anything to simulate a timeout
				}()
				return func() { ln.Close() }
			},
			wantBanner: "",
			wantErr:    "failed to read ssh-banner",
		},
		{
			name:    "failure to read banner line",
			address: "127.0.0.1:2224",
			setupServer: func(address string) func() {
				ln, err := net.Listen("tcp", address)
				require.NoError(t, err)
				go func() {
					conn, _ := ln.Accept()
					defer conn.Close()
					_, _ = conn.Write([]byte("Partial banner"))
					// No newline to simulate an incomplete banner
				}()
				return func() { ln.Close() }
			},
			wantBanner: "",
			wantErr:    "failed to read ssh-banner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown := tt.setupServer(tt.address)
			defer teardown()

			sbr := &defaultSshBannerReader{}

			banner, err := sbr.readSSHBanner(tt.address)
			if tt.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, tt.wantBanner, banner)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}
