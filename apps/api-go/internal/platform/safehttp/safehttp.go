package safehttp

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ErrUnsafeDestination is returned when an outbound URL violates SSRF or allowlist policy.
var ErrUnsafeDestination = errors.New("invalid or unsupported destination URL")

// Options configures allowlist and private-network SSRF checks.
type Options struct {
	Allowlist    []string
	AllowPrivate bool
	OriginOnly   bool
	Timeout      time.Duration
	LookupIPs    func(ctx context.Context, host string) ([]netip.Addr, error)
}

// Destination represents a validated outbound target bound to a resolved IP-pinned HTTP client.
type Destination struct {
	Origin   string
	URL      string
	Hostname string
	PinnedIP string
	Client   *http.Client
}

// ParseHostAllowlist normalizes comma/entry allowlist entries or applies defaults.
func ParseHostAllowlist(raw, fallback []string) []string {
	out := make([]string, 0, len(raw))
	for _, entry := range raw {
		for _, part := range strings.Split(entry, ",") {
			trimmed := strings.ToLower(strings.TrimSpace(part))
			if trimmed != "" {
				out = append(out, trimmed)
			}
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}

// HostnameAllowed checks exact matches, '*', and single-label '*.domain' wildcards.
func HostnameAllowed(hostname string, allowlist []string) bool {
	host := strings.ToLower(strings.TrimSpace(hostname))
	if host == "" {
		return false
	}
	for _, pattern := range allowlist {
		p := strings.ToLower(strings.TrimSpace(pattern))
		switch {
		case p == "*":
			return true
		case strings.HasPrefix(p, "*."):
			suffix := p[1:] // ".atlassian.net"
			if strings.HasSuffix(host, suffix) && strings.IndexByte(host, '.') == len(host)-len(suffix) {
				return true
			}
		case p == host:
			return true
		}
	}
	return false
}

// IsPrivateIP returns true for loopback, private, link-local, CGNAT, multicast, and non-global addresses.
func IsPrivateIP(addr netip.Addr) bool {
	if !addr.IsValid() {
		return true
	}
	addr = addr.Unmap()
	if addr.IsLoopback() || addr.IsPrivate() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() || addr.IsMulticast() || addr.IsUnspecified() {
		return true
	}
	if addr.Is4() {
		b := addr.As4()
		switch {
		case b[0] == 0:
			return true
		case b[0] == 100 && b[1] >= 64 && b[1] <= 127: // CGNAT 100.64.0.0/10
			return true
		case b[0] == 192 && b[1] == 0 && (b[2] == 0 || b[2] == 2): // IETF / TEST-NET-1
			return true
		case b[0] == 198 && (b[1] == 18 || b[1] == 19): // Benchmarking
			return true
		case b[0] == 198 && b[1] == 51 && b[2] == 100: // TEST-NET-2
			return true
		case b[0] == 203 && b[1] == 0 && b[2] == 113: // TEST-NET-3
			return true
		case b[0] >= 224: // Multicast / reserved
			return true
		}
	}
	return false
}

// ResolveSafeDestination validates a URL against SSRF policy and pins dials to the verified IP.
func ResolveSafeDestination(ctx context.Context, rawURL string, opts Options) (*Destination, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("%w: malformed URL", ErrUnsafeDestination)
	}

	isHTTPS := strings.EqualFold(parsed.Scheme, "https")
	isHTTP := strings.EqualFold(parsed.Scheme, "http")
	if !isHTTPS && !(isHTTP && opts.AllowPrivate) {
		return nil, fmt.Errorf("%w: protocol not permitted", ErrUnsafeDestination)
	}
	if parsed.User != nil {
		return nil, fmt.Errorf("%w: URL must not include credentials", ErrUnsafeDestination)
	}
	if opts.OriginOnly {
		if parsed.RawQuery != "" || parsed.Fragment != "" {
			return nil, fmt.Errorf("%w: URL must not include query or fragment", ErrUnsafeDestination)
		}
		if parsed.Path != "" && parsed.Path != "/" {
			return nil, fmt.Errorf("%w: URL must be an origin without path", ErrUnsafeDestination)
		}
	}

	hostname := strings.ToLower(parsed.Hostname())
	if !HostnameAllowed(hostname, opts.Allowlist) {
		return nil, fmt.Errorf("%w: hostname not allowlisted", ErrUnsafeDestination)
	}

	ips, err := lookupAddrs(ctx, hostname, opts.LookupIPs)
	if err != nil || len(ips) == 0 {
		return nil, fmt.Errorf("%w: hostname resolution failed", ErrUnsafeDestination)
	}
	if !opts.AllowPrivate {
		for _, ip := range ips {
			if IsPrivateIP(ip) {
				return nil, fmt.Errorf("%w: non-public IP refused", ErrUnsafeDestination)
			}
		}
	}

	portNum, err := resolvePort(parsed, isHTTPS)
	if err != nil {
		return nil, err
	}

	pinnedIP := ips[0].Unmap().String()
	client := buildPinnedClient(pinnedIP, portNum, hostname, isHTTPS, opts.Timeout)
	origin := fmt.Sprintf("%s://%s", strings.ToLower(parsed.Scheme), parsed.Host)
	cleanURL := origin
	if !opts.OriginOnly {
		cleanURL = parsed.String()
	}

	return &Destination{
		Origin:   origin,
		URL:      cleanURL,
		Hostname: hostname,
		PinnedIP: pinnedIP,
		Client:   client,
	}, nil
}

func lookupAddrs(ctx context.Context, hostname string, fn func(context.Context, string) ([]netip.Addr, error)) ([]netip.Addr, error) {
	if fn != nil {
		return fn(ctx, hostname)
	}
	if ip, err := netip.ParseAddr(hostname); err == nil {
		return []netip.Addr{ip}, nil
	}
	resolved, err := net.DefaultResolver.LookupIPAddr(ctx, hostname)
	if err != nil {
		return nil, err
	}
	out := make([]netip.Addr, 0, len(resolved))
	for _, r := range resolved {
		if addr, ok := netip.AddrFromSlice(r.IP); ok {
			out = append(out, addr.Unmap())
		}
	}
	return out, nil
}

func resolvePort(parsed *url.URL, isHTTPS bool) (int, error) {
	if rawPort := parsed.Port(); rawPort != "" {
		p, err := strconv.Atoi(rawPort)
		if err != nil || p < 1 || p > 65535 {
			return 0, fmt.Errorf("%w: invalid port", ErrUnsafeDestination)
		}
		return p, nil
	}
	if isHTTPS {
		return 443, nil
	}
	return 80, nil
}

func buildPinnedClient(pinnedIP string, port int, tlsServerName string, isHTTPS bool, timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	pinnedAddr := net.JoinHostPort(pinnedIP, strconv.Itoa(port))
	dialer := &net.Dialer{Timeout: 10 * time.Second}

	tr := &http.Transport{
		DisableKeepAlives: true,
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, pinnedAddr)
		},
	}
	if isHTTPS {
		tr.TLSClientConfig = &tls.Config{
			ServerName: tlsServerName,
			MinVersion: tls.VersionTLS12,
		}
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: tr,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
