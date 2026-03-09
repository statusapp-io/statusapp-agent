package checks

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

// RunSSL performs an SSL certificate check — verifies the cert is valid and not expiring soon.
func RunSSL(m Monitor) Result {
	host := m.URL
	for _, prefix := range []string{"https://", "http://", "ssl://"} {
		if len(host) > len(prefix) && host[:len(prefix)] == prefix {
			host = host[len(prefix):]
			break
		}
	}
	for i, c := range host {
		if c == '/' {
			host = host[:i]
			break
		}
	}

	// Strip port if present, default to 443
	port := "443"
	if h, p, err := net.SplitHostPort(host); err == nil {
		host = h
		port = p
	}

	addr := net.JoinHostPort(host, port)

	timeout := time.Duration(m.Timeout) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	start := time.Now()

	dialer := &net.Dialer{Timeout: timeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{
		ServerName: host,
	})
	responseTime := int(time.Since(start).Milliseconds())

	if err != nil {
		return Result{
			Status:       "DOWN",
			ResponseTime: &responseTime,
			Error:        fmt.Sprintf("TLS handshake failed: %v", err),
		}
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return Result{
			Status:       "DOWN",
			ResponseTime: &responseTime,
			Error:        "no certificates presented",
		}
	}

	leaf := certs[0]
	now := time.Now()

	// Check if expired
	if now.After(leaf.NotAfter) {
		return Result{
			Status:       "DOWN",
			ResponseTime: &responseTime,
			Error:        fmt.Sprintf("certificate expired on %s", leaf.NotAfter.Format("2006-01-02")),
		}
	}

	// Check if not yet valid
	if now.Before(leaf.NotBefore) {
		return Result{
			Status:       "DOWN",
			ResponseTime: &responseTime,
			Error:        fmt.Sprintf("certificate not valid until %s", leaf.NotBefore.Format("2006-01-02")),
		}
	}

	// Check expiry warning threshold
	expiryDays := 30
	if m.SSLCertExpiryDays != nil && *m.SSLCertExpiryDays > 0 {
		expiryDays = *m.SSLCertExpiryDays
	}
	daysUntilExpiry := int(leaf.NotAfter.Sub(now).Hours() / 24)

	if daysUntilExpiry <= expiryDays {
		return Result{
			Status:       "DEGRADED",
			ResponseTime: &responseTime,
			Error:        fmt.Sprintf("certificate expires in %d days (threshold: %d)", daysUntilExpiry, expiryDays),
		}
	}

	return Result{
		Status:       "UP",
		ResponseTime: &responseTime,
	}
}
