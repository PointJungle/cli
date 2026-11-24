package cli

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
)

// fakeServer plays the PointJungle API and its identity provider on one
// server: /api/cli/config, discovery, authorize (with PKCE), token
// (authorization code, refresh token, device code), device authorization,
// revocation and /api/profiles/me/.
type fakeServer struct {
	t      *testing.T
	server *httptest.Server

	mu            sync.Mutex
	codes         map[string]pendingCode // authorization code → request
	access        map[string]bool        // valid access tokens
	refresh       map[string]bool        // valid refresh tokens
	revoked       []string
	refreshCalls  int
	refuseRefresh bool
	deviceAllowed bool
	denyAuthorize bool
	expiresIn     int
	noSignedInURL bool     // an older host: no signed_in_url in /api/cli/config
	prompts       []string // prompt of each authorization request
	landings      []string // query of each visit to /cli/signed-in
}

type pendingCode struct {
	challenge   string
	redirectURI string
}

const clientID = "codeforge-cli"

func newFake(t *testing.T) *fakeServer {
	f := &fakeServer{
		t:         t,
		codes:     map[string]pendingCode{},
		access:    map[string]bool{},
		refresh:   map[string]bool{},
		expiresIn: 3600,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/cli/config", f.cliConfig)
	mux.HandleFunc("GET /.well-known/openid-configuration", f.discovery)
	mux.HandleFunc("GET /authorize", f.authorize)
	mux.HandleFunc("POST /token", f.token)
	mux.HandleFunc("POST /device", f.device)
	mux.HandleFunc("POST /revoke", f.revoke)
	mux.HandleFunc("GET /api/profiles/me/", f.me)
	mux.HandleFunc("GET /cli/signed-in", f.signedIn)
	f.server = httptest.NewServer(mux)
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakeServer) URL() string { return f.server.URL }

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func randomToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (f *fakeServer) cliConfig(w http.ResponseWriter, _ *http.Request) {
	cfg := map[string]any{
		"api_url":      f.URL(),
		"issuer":       f.URL(),
		"client_id":    clientID,
		"scopes":       []string{"openid", "email", "profile"},
		"redirect_uri": "http://127.0.0.1/callback",
	}
	if !f.noSignedInURL {
		cfg["signed_in_url"] = f.URL() + "/cli/signed-in"
	}
	writeJSON(w, 200, cfg)
}

func (f *fakeServer) signedIn(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	f.landings = append(f.landings, r.URL.RawQuery)
	f.mu.Unlock()
	_, _ = w.Write([]byte("<h1>Signed in</h1>"))
}

func (f *fakeServer) discovery(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]string{
		"issuer":                        f.URL(),
		"authorization_endpoint":        f.URL() + "/authorize",
		"token_endpoint":                f.URL() + "/token",
		"device_authorization_endpoint": f.URL() + "/device",
		"revocation_endpoint":           f.URL() + "/revoke",
	})
}

func (f *fakeServer) authorize(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	redirect, err := url.Parse(q.Get("redirect_uri"))
	if err != nil || redirect.Hostname() != "127.0.0.1" || redirect.Path != "/callback" || redirect.Port() == "" {
		http.Error(w, "redirect_uri must be http://127.0.0.1:<port>/callback", 400)
		return
	}
	if q.Get("client_id") != clientID || q.Get("response_type") != "code" ||
		q.Get("code_challenge_method") != "S256" || q.Get("code_challenge") == "" {
		http.Error(w, "bad authorization request", 400)
		return
	}
	f.mu.Lock()
	f.prompts = append(f.prompts, q.Get("prompt"))
	f.mu.Unlock()
	target := redirect.Query()
	target.Set("state", q.Get("state"))
	if f.denyAuthorize {
		target.Set("error", "access_denied")
	} else {
		code := randomToken()
		f.mu.Lock()
		f.codes[code] = pendingCode{challenge: q.Get("code_challenge"), redirectURI: redirect.String()}
		f.mu.Unlock()
		target.Set("code", code)
	}
	redirect.RawQuery = target.Encode()
	http.Redirect(w, r, redirect.String(), http.StatusFound)
}

func (f *fakeServer) issue(w http.ResponseWriter) {
	access, refresh := randomToken(), randomToken()
	f.access[access] = true
	f.refresh[refresh] = true
	writeJSON(w, 200, map[string]any{
		"access_token":  access,
		"refresh_token": refresh,
		"token_type":    "Bearer",
		"expires_in":    f.expiresIn,
		"id_token":      "header.payload.signature",
	})
}

func (f *fakeServer) token(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	f.mu.Lock()
	defer f.mu.Unlock()
	if r.PostForm.Get("client_id") != clientID {
		writeJSON(w, 401, map[string]string{"error": "invalid_client"})
		return
	}
	switch r.PostForm.Get("grant_type") {
	case "authorization_code":
		pending, ok := f.codes[r.PostForm.Get("code")]
		delete(f.codes, r.PostForm.Get("code"))
		sum := sha256.Sum256([]byte(r.PostForm.Get("code_verifier")))
		if !ok || pending.redirectURI != r.PostForm.Get("redirect_uri") ||
			base64.RawURLEncoding.EncodeToString(sum[:]) != pending.challenge {
			writeJSON(w, 400, map[string]string{"error": "invalid_grant"})
			return
		}
		f.issue(w)
	case "refresh_token":
		f.refreshCalls++
		old := r.PostForm.Get("refresh_token")
		if f.refuseRefresh || !f.refresh[old] {
			writeJSON(w, 400, map[string]string{"error": "invalid_grant"})
			return
		}
		delete(f.refresh, old) // rotation
		f.issue(w)
	case "urn:ietf:params:oauth:grant-type:device_code":
		if r.PostForm.Get("device_code") != "device-123" {
			writeJSON(w, 400, map[string]string{"error": "invalid_grant"})
			return
		}
		f.issue(w)
	default:
		writeJSON(w, 400, map[string]string{"error": "unsupported_grant_type"})
	}
}

func (f *fakeServer) device(w http.ResponseWriter, _ *http.Request) {
	if !f.deviceAllowed {
		writeJSON(w, 400, map[string]string{"error": "unauthorized_client"})
		return
	}
	writeJSON(w, 200, map[string]any{
		"device_code":      "device-123",
		"user_code":        "ABCD-EFGH",
		"verification_uri": f.URL() + "/device/verify",
		"expires_in":       600,
		"interval":         1,
	})
}

func (f *fakeServer) revoke(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	f.mu.Lock()
	defer f.mu.Unlock()
	token := r.PostForm.Get("token")
	f.revoked = append(f.revoked, token)
	delete(f.refresh, token)
	w.WriteHeader(200)
}

func (f *fakeServer) me(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	f.mu.Lock()
	valid := f.access[token]
	f.mu.Unlock()
	if !valid {
		writeJSON(w, 401, map[string]string{"code": "token_invalid"})
		return
	}
	writeJSON(w, 200, map[string]any{"user": map[string]any{
		"id": "u-1", "email": "ada@example.com", "display_name": "Ada Lovelace",
	}})
}

// invalidateAccessTokens makes the API refuse every current access token.
func (f *fakeServer) invalidateAccessTokens() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.access = map[string]bool{}
}

// browse plays the person in the browser: follows the authorization URL to the
// CLI's loopback redirect.
func browse(t *testing.T) func(string) error {
	return func(authURL string) error {
		resp, err := http.Get(authURL)
		if err != nil {
			t.Errorf("browser: %v", err)
			return err
		}
		resp.Body.Close()
		return nil
	}
}
