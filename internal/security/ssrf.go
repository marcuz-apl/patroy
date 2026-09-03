package security

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// ValidateTargetURL verifies that a target URL is syntactically valid, uses HTTP/HTTPS,
// and does not target internal loopback, private networks, or cloud metadata endpoints
// when allowPrivate is false.
func ValidateTargetURL(rawURL string, allowPrivate bool) error {
	if strings.TrimSpace(rawURL) == "" {
		return fmt.Errorf("security: empty target URL")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("security: invalid URL syntax: %w", err)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("security: unsupported scheme %q (only http and https are allowed)", parsed.Scheme)
	}

	host := parsed.Hostname()
	if host == "" {
		return fmt.Errorf("security: missing hostname in target URL")
	}

	if allowPrivate {
		return nil
	}

	// Fast check for common localhost names
	lowerHost := strings.ToLower(host)
	if lowerHost == "localhost" || strings.HasSuffix(lowerHost, ".localhost") || strings.HasSuffix(lowerHost, ".local") {
		return fmt.Errorf("security: access to %q is restricted (SSRF protection)", host)
	}

	// Check direct IP or resolve domain
	ips, err := net.LookupIP(host)
	if err != nil {
		// If DNS resolution fails here, let the browser or HTTP client handle the connection failure
		return nil
	}

	for _, ip := range ips {
		if isRestrictedIP(ip) {
			return fmt.Errorf("security: target %q resolves to restricted IP %s (SSRF protection)", host, ip.String())
		}
	}

	return nil
}

// isRestrictedIP checks if an IP belongs to loopback, private, link-local, or cloud metadata ranges.
func isRestrictedIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return true
	}

	// Check IPv4 link-local (169.254.0.0/16 including 169.254.169.254 cloud metadata)
	if ipv4 := ip.To4(); ipv4 != nil {
		if ipv4[0] == 169 && ipv4[1] == 254 {
			return true
		}
		// 127.0.0.0/8 loopback
		if ipv4[0] == 127 {
			return true
		}
		// 0.0.0.0/8
		if ipv4[0] == 0 {
			return true
		}
	}

	return false
}
