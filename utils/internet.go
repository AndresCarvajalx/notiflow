package utils

import (
	"context"
	"net"
	"time"
)

func WaitForInternet(ctx context.Context) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Attempt to dial a public DNS server
			conn, err := net.DialTimeout("tcp", "8.8.8.8:53", 1*time.Second)
			if err == nil {
				conn.Close()
				return nil
			}
		}
	}
}
