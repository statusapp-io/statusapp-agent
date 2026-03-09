package checks

import (
	"fmt"
	"net"
	"time"
)

// RunPing performs an ICMP-like connectivity check.
// Uses TCP connect to port 80 as a fallback since raw ICMP requires root.
// If the monitor has a port configured, use that instead.
func RunPing(m Monitor) Result {
	host := m.URL
	for _, prefix := range []string{"https://", "http://", "ping://", "icmp://"} {
		if len(host) > len(prefix) && host[:len(prefix)] == prefix {
			host = host[len(prefix):]
			break
		}
	}
	for i, c := range host {
		if c == '/' || c == ':' {
			host = host[:i]
			break
		}
	}

	// Resolve hostname first
	start := time.Now()
	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		responseTime := int(time.Since(start).Milliseconds())
		return Result{
			Status:       "DOWN",
			ResponseTime: &responseTime,
			Error:        fmt.Sprintf("DNS resolution failed for %s: %v", host, err),
		}
	}

	ip := ips[0].String()

	timeout := time.Duration(m.Timeout) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	// Try TCP connect as a connectivity check (ICMP requires root)
	port := 80
	if m.Port != nil && *m.Port > 0 {
		port = *m.Port
	}

	addr := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	responseTime := int(time.Since(start).Milliseconds())

	if err != nil {
		return Result{
			Status:       "DOWN",
			ResponseTime: &responseTime,
			Error:        fmt.Sprintf("host %s (%s) unreachable: %v", host, ip, err),
		}
	}
	conn.Close()

	return Result{
		Status:       "UP",
		ResponseTime: &responseTime,
	}
}
