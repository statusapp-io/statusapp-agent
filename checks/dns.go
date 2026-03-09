package checks

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// RunDNS performs a DNS resolution check.
func RunDNS(m Monitor) Result {
	// Extract hostname from URL
	host := m.URL
	for _, prefix := range []string{"https://", "http://", "dns://"} {
		if len(host) > len(prefix) && host[:len(prefix)] == prefix {
			host = host[len(prefix):]
			break
		}
	}
	// Strip trailing path/port
	for i, c := range host {
		if c == '/' || c == ':' {
			host = host[:i]
			break
		}
	}

	start := time.Now()

	recordType := "A"
	if m.DNSRecord != nil && *m.DNSRecord != "" {
		recordType = strings.ToUpper(*m.DNSRecord)
	}

	var records []string
	var err error

	switch recordType {
	case "A", "AAAA":
		var ips []net.IP
		ips, err = net.LookupIP(host)
		if err == nil {
			for _, ip := range ips {
				if recordType == "A" && ip.To4() != nil {
					records = append(records, ip.String())
				} else if recordType == "AAAA" && ip.To4() == nil {
					records = append(records, ip.String())
				}
			}
		}
	case "MX":
		var mxs []*net.MX
		mxs, err = net.LookupMX(host)
		if err == nil {
			for _, mx := range mxs {
				records = append(records, mx.Host)
			}
		}
	case "TXT":
		records, err = net.LookupTXT(host)
	case "NS":
		var nss []*net.NS
		nss, err = net.LookupNS(host)
		if err == nil {
			for _, ns := range nss {
				records = append(records, ns.Host)
			}
		}
	case "CNAME":
		var cname string
		cname, err = net.LookupCNAME(host)
		if err == nil {
			records = append(records, cname)
		}
	default:
		var ips []net.IP
		ips, err = net.LookupIP(host)
		if err == nil {
			for _, ip := range ips {
				records = append(records, ip.String())
			}
		}
	}

	responseTime := int(time.Since(start).Milliseconds())

	if err != nil {
		return Result{
			Status:       "DOWN",
			ResponseTime: &responseTime,
			Error:        fmt.Sprintf("DNS lookup failed for %s: %v", host, err),
		}
	}

	if len(records) == 0 {
		return Result{
			Status:       "DOWN",
			ResponseTime: &responseTime,
			Error:        fmt.Sprintf("no %s records found for %s", recordType, host),
		}
	}

	// Check expected IP if configured
	if m.ExpectedIP != nil && *m.ExpectedIP != "" {
		found := false
		for _, r := range records {
			if r == *m.ExpectedIP {
				found = true
				break
			}
		}
		if !found {
			return Result{
				Status:       "DOWN",
				ResponseTime: &responseTime,
				Error:        fmt.Sprintf("expected %s but got %s", *m.ExpectedIP, strings.Join(records, ", ")),
			}
		}
	}

	return Result{
		Status:       "UP",
		ResponseTime: &responseTime,
	}
}
