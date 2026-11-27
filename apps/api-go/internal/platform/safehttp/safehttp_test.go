package safehttp

import (
	"context"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostnameAllowedAndPrivateIP(t *testing.T) {
	allow := ParseHostAllowlist([]string{"api.openai.com,*.anthropic.com"}, nil)
	assert.True(t, HostnameAllowed("api.openai.com", allow))
	assert.True(t, HostnameAllowed("api.anthropic.com", allow))
	assert.False(t, HostnameAllowed("a.b.anthropic.com", allow), "single-label wildcard only")
	assert.False(t, HostnameAllowed("evil.com", allow))

	assert.True(t, IsPrivateIP(netip.MustParseAddr("127.0.0.1")))
	assert.True(t, IsPrivateIP(netip.MustParseAddr("169.254.169.254")))
	assert.True(t, IsPrivateIP(netip.MustParseAddr("192.168.1.50")))
	assert.True(t, IsPrivateIP(netip.MustParseAddr("100.64.0.1")))
	assert.True(t, IsPrivateIP(netip.MustParseAddr("::1")))
	assert.False(t, IsPrivateIP(netip.MustParseAddr("93.184.216.34")))
}

func TestResolveSafeDestinationSSRFBlocks(t *testing.T) {
	ctx := context.Background()

	// Reject HTTP when private endpoints are disallowed.
	_, err := ResolveSafeDestination(ctx, "http://api.openai.com/v1", Options{
		Allowlist: []string{"api.openai.com"},
		LookupIPs: func(context.Context, string) ([]netip.Addr, error) {
			return []netip.Addr{netip.MustParseAddr("93.184.216.34")}, nil
		},
	})
	require.ErrorIs(t, err, ErrUnsafeDestination)

	// Reject mixed DNS containing a private IP.
	_, err = ResolveSafeDestination(ctx, "https://api.openai.com/v1", Options{
		Allowlist: []string{"api.openai.com"},
		LookupIPs: func(context.Context, string) ([]netip.Addr, error) {
			return []netip.Addr{
				netip.MustParseAddr("93.184.216.34"),
				netip.MustParseAddr("10.0.0.8"),
			}, nil
		},
	})
	require.ErrorIs(t, err, ErrUnsafeDestination)

	// Allow private when explicitly configured.
	dest, err := ResolveSafeDestination(ctx, "http://localhost:11434/v1", Options{
		Allowlist:    []string{"*"},
		AllowPrivate: true,
		LookupIPs: func(context.Context, string) ([]netip.Addr, error) {
			return []netip.Addr{netip.MustParseAddr("127.0.0.1")}, nil
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1", dest.PinnedIP)
	assert.Equal(t, "localhost", dest.Hostname)
}
