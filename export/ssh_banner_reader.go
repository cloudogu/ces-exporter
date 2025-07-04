package export

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"time"
)

const dialReadTimeoutForDoguSidecar = 2 * time.Second

type defaultSshBannerReader struct {
}

func (sbr *defaultSshBannerReader) readSSHBanner(address string) (string, error) {
	conn, err := net.DialTimeout("tcp", address, dialReadTimeoutForDoguSidecar)
	if err != nil {
		return "", fmt.Errorf("connection failed while getting ssh-banner: %w", err)
	}

	if err := conn.SetReadDeadline(time.Now().Add(dialReadTimeoutForDoguSidecar)); err != nil {
		return "", fmt.Errorf("failed to set read-deadline while getting ssh-banner: %w", err)
	}
	defer func() {
		if cErr := conn.Close(); cErr != nil {
			slog.Warn("Failed to close connection getting ssh-banner", "error", cErr)
		}
	}()

	reader := bufio.NewReader(conn)
	banner, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read ssh-banner: %w", err)
	}

	return banner, nil
}
