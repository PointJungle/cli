package httpclient

import (
	"net"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"
)

func TestLocalhostAddr(t *testing.T) {
	for addr, want := range map[string]string{
		"accounts.localhost:5173": "localhost:5173",
		"A.B.LOCALHOST.:80":       "localhost:80",
		"localhost:8000":          "localhost:8000",
		"127.0.0.1:8000":          "127.0.0.1:8000",
		"api.example.com:443":     "api.example.com:443",
		"localhost.example.com:1": "localhost.example.com:1",
		"notlocalhost:1":          "notlocalhost:1",
	} {
		if got := LocalhostAddr(addr); got != want {
			t.Errorf("LocalhostAddr(%q) = %q, want %q", addr, got, want)
		}
	}
}

func TestClientReachesLocalhostSubdomainAndKeepsHost(t *testing.T) {
	var gotHost, gotAgent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHost, gotAgent = r.Host, r.UserAgent()
	}))
	defer server.Close()
	_, port, _ := net.SplitHostPort(server.Listener.Addr().String())

	resp, err := New(5*time.Second, UserAgent("1.2.3")).Get("http://accounts.localhost:" + port + "/")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	resp.Body.Close()
	if gotHost != "accounts.localhost:"+port {
		t.Fatalf("Host header = %q", gotHost)
	}
	if want := "pj-cli/1.2.3 (" + runtime.GOOS + "; " + runtime.GOARCH + ")"; gotAgent != want {
		t.Fatalf("User-Agent = %q, want %q", gotAgent, want)
	}
}
