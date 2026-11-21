// Package httpclient builds the CLI's HTTP client.
package httpclient

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"
)

// New returns an HTTP client that, like browsers, sends every *.localhost name
// to this machine (RFC 6761): a local codeforge serves its sign-in pages on
// names such as accounts.localhost, which DNS servers do not know.
func New(timeout time.Duration) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	dialer := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, LocalhostAddr(addr))
	}
	return &http.Client{Timeout: timeout, Transport: transport}
}

// LocalhostAddr maps "name.localhost:port" to "localhost:port" (resolved by
// the system to 127.0.0.1 or ::1) and leaves any other address unchanged. Only
// the connection changes: the request keeps its Host header.
func LocalhostAddr(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil || !IsLocalhostName(host) || strings.EqualFold(host, "localhost") {
		return addr
	}
	return net.JoinHostPort("localhost", port)
}

// IsLocalhostName reports whether host is "localhost" or a name under it.
func IsLocalhostName(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	return host == "localhost" || strings.HasSuffix(host, ".localhost")
}
