package checks

import (
	"fmt"
	"net"
	"time"
)

// RunTCP performs a TCP port connectivity check.
func RunTCP(m Monitor) Result {
	port := 80
	if m.Port != nil && *m.Port > 0 {
		port = *m.Port
	}

	// Extract host from URL, stripping protocol
	host := m.URL
	for _, prefix := range []string{"https://", "http://", "tcp://"} {
		if len(host) > len(prefix) && host[:len(prefix)] == prefix {
			host = host[len(prefix):]
			break
		}
	}
	// Strip trailing path
	for i, c := range host {
		if c == '/' {
			host = host[:i]
			break
		}
	}
	// Strip existing port if present
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))

	timeout := time.Duration(m.Timeout) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, timeout)
	responseTime := int(time.Since(start).Milliseconds())

	if err != nil {
		return Result{
			Status:       "DOWN",
			ResponseTime: &responseTime,
			Error:        fmt.Sprintf("TCP connection to %s failed: %v", addr, err),
		}
	}
	conn.Close()

	return Result{
		Status:       "UP",
		ResponseTime: &responseTime,
	}
}
