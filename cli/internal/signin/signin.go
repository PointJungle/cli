// Package signin signs a person in to codeforge through its OpenID Connect
// identity provider, and keeps the resulting tokens fresh.
//
// The host tells the CLI what it needs (GET <host>/api/cli/config: the API URL,
// the issuer, the CLI's client ID and scopes, its redirect URI, the page to
// show once signed in); the issuer's
// discovery document (<issuer>/.well-known/openid-configuration) gives the
// endpoints. The CLI is a public client: it has no secret and proves the sign-in
// with PKCE.
//
// Two ways to sign in:
//
//   - Loopback (default): the CLI listens on 127.0.0.1 on a random port, opens
//     the browser on the authorization endpoint (prompt=select_account, so the
//     person picks the account), receives the authorization code on
//     http://127.0.0.1:<port>/callback, exchanges it for tokens, then sends the
//     browser to the host's signed-in page (?error=<code> on failure).
//   - Device: the CLI shows a code and a URL to open on any device, then polls
//     the token endpoint until the person has approved — for machines without a
//     browser. The host must allow the device grant for the CLI's client.
package signin

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/oauth2"
)

// ErrDeviceUnavailable means the host does not allow the device sign-in.
var ErrDeviceUnavailable = errors.New("device sign-in is not available on this host")

// CLIConfig is the response of GET <host>/api/cli/config.
type CLIConfig struct {
	APIURL      string   `json:"api_url"`
	Issuer      string   `json:"issuer"`
	ClientID    string   `json:"client_id"`
	Scopes      []string `json:"scopes"`
	RedirectURI string   `json:"redirect_uri"`
	// SignedInURL is the page the browser is sent to once the sign-in is done,
	// with ?error=<code> when it failed. Empty on older hosts: the CLI then
	// shows its own page.
	SignedInURL string `json:"signed_in_url"`
}

// Discovery is the part of the issuer's discovery document the CLI uses.
type Discovery struct {
	AuthorizationEndpoint       string `json:"authorization_endpoint"`
	TokenEndpoint               string `json:"token_endpoint"`
	DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint"`
	RevocationEndpoint          string `json:"revocation_endpoint"`
}

// FetchCLIConfig reads what the host's CLI needs.
func FetchCLIConfig(ctx context.Context, client *http.Client, host string) (*CLIConfig, error) {
	var cfg CLIConfig
	if err := getJSON(ctx, client, strings.TrimRight(host, "/")+"/api/cli/config", &cfg); err != nil {
		return nil, fmt.Errorf("reading the CLI configuration of %s: %w", host, err)
	}
	if cfg.Issuer == "" || cfg.ClientID == "" || cfg.APIURL == "" {
		return nil, fmt.Errorf("incomplete CLI configuration from %s", host)
	}
	return &cfg, nil
}

// Discover reads the issuer's discovery document.
func Discover(ctx context.Context, client *http.Client, issuer string) (*Discovery, error) {
	var d Discovery
	endpoint := strings.TrimRight(issuer, "/") + "/.well-known/openid-configuration"
	if err := getJSON(ctx, client, endpoint, &d); err != nil {
		return nil, fmt.Errorf("reading the identity provider's configuration: %w", err)
	}
	if d.AuthorizationEndpoint == "" || d.TokenEndpoint == "" {
		return nil, errors.New("the identity provider's configuration has no authorization or token endpoint")
	}
	return &d, nil
}

// Flow signs a person in.
type Flow struct {
	HTTP *http.Client
	// OpenBrowser opens a URL; its failure is not fatal (the URL is printed).
	OpenBrowser func(string) error
	// Out receives the messages for the person (stderr).
	Out io.Writer
}

func (f Flow) oauthConfig(cfg *CLIConfig, d *Discovery, redirect string) *oauth2.Config {
	return &oauth2.Config{
		ClientID: cfg.ClientID,
		Endpoint: oauth2.Endpoint{
			AuthURL:       d.AuthorizationEndpoint,
			TokenURL:      d.TokenEndpoint,
			DeviceAuthURL: d.DeviceAuthorizationEndpoint,
			AuthStyle:     oauth2.AuthStyleInParams,
		},
		RedirectURL: redirect,
		Scopes:      cfg.Scopes,
	}
}

func (f Flow) withClient(ctx context.Context) context.Context {
	return context.WithValue(ctx, oauth2.HTTPClient, f.HTTP)
}

type callback struct {
	code    string // authorization code
	errCode string // OAuth-style error code for the browser page
	err     error
}

type loopbackResult struct {
	token *oauth2.Token
	err   error
}

// Loopback signs in through the browser and a local redirect.
func (f Flow) Loopback(ctx context.Context, cfg *CLIConfig, d *Discovery) (*oauth2.Token, error) {
	redirect, err := url.Parse(cfg.RedirectURI)
	if err != nil || redirect.Hostname() == "" {
		return nil, fmt.Errorf("invalid redirect URI %q", cfg.RedirectURI)
	}
	listener, err := net.Listen("tcp", net.JoinHostPort(redirect.Hostname(), "0"))
	if err != nil {
		return nil, fmt.Errorf("listening for the sign-in redirect: %w", err)
	}
	defer listener.Close()
	redirect.Host = listener.Addr().String()

	oc := f.oauthConfig(cfg, d, redirect.String())
	state := randomString()
	verifier := oauth2.GenerateVerifier()
	// prompt=select_account: the account site shows its account chooser
	// instead of signing in with whichever account is active there.
	authURL := oc.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier),
		oauth2.SetAuthURLParam("prompt", "select_account"))

	// The code is exchanged before answering the browser, so that its page
	// tells the truth about the sign-in.
	results := make(chan loopbackResult, 1)
	var handled atomic.Bool
	server := &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != redirect.Path {
				http.NotFound(w, r)
				return
			}
			if handled.Swap(true) { // a reload of the callback: already done
				finishInBrowser(w, r, cfg.SignedInURL, "", nil)
				return
			}
			cb := readCallback(r.URL.Query(), state)
			result := loopbackResult{err: cb.err}
			if cb.err == nil {
				token, err := oc.Exchange(f.withClient(ctx), cb.code, oauth2.VerifierOption(verifier))
				if err != nil {
					cb.errCode = "token_exchange_failed"
					result.err = fmt.Errorf("exchanging the authorization code: %w", describe(err))
				}
				result.token = token
			}
			finishInBrowser(w, r, cfg.SignedInURL, cb.errCode, result.err)
			results <- result
		}),
	}
	go func() { _ = server.Serve(listener) }()
	defer server.Close()

	if f.OpenBrowser != nil {
		fmt.Fprintf(f.Out, "Opening your browser to sign in. If it does not open, visit:\n\n  %s\n\n", authURL)
		_ = f.OpenBrowser(authURL)
	} else {
		fmt.Fprintf(f.Out, "To sign in, open this URL in a browser on this machine:\n\n  %s\n\n", authURL)
	}

	select {
	case result := <-results:
		return result.token, result.err
	case <-ctx.Done():
		return nil, fmt.Errorf("sign-in not completed: %w", ctx.Err())
	}
}

func readCallback(query url.Values, state string) callback {
	if e := query.Get("error"); e != "" {
		msg := e
		if desc := query.Get("error_description"); desc != "" {
			msg += ": " + desc
		}
		return callback{errCode: errorCode(e), err: fmt.Errorf("sign-in refused: %s", msg)}
	}
	if query.Get("state") != state {
		return callback{errCode: "invalid_state", err: errors.New("sign-in redirect with an unexpected state; try again")}
	}
	code := query.Get("code")
	if code == "" {
		return callback{errCode: "invalid_request", err: errors.New("sign-in redirect without an authorization code")}
	}
	return callback{code: code}
}

// errorCode keeps an OAuth error code safe to put in a URL (letters and _).
func errorCode(e string) string {
	for _, c := range e {
		if (c < 'a' || c > 'z') && c != '_' {
			return "access_denied"
		}
	}
	return e
}

// finishInBrowser sends the browser to the host's signed-in page, with
// ?error=<code> when the sign-in failed, or shows the CLI's own page when the
// host gives none.
func finishInBrowser(w http.ResponseWriter, r *http.Request, signedInURL, errCode string, err error) {
	if target, ok := signedInTarget(signedInURL, errCode, err); ok {
		http.Redirect(w, r, target, http.StatusFound)
		return
	}
	writeCallbackPage(w, err)
}

func signedInTarget(raw, errCode string, err error) (string, bool) {
	u, parseErr := url.Parse(raw)
	if raw == "" || parseErr != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return "", false
	}
	if err != nil {
		if errCode == "" {
			errCode = "sign_in_failed"
		}
		q := u.Query()
		q.Set("error", errCode)
		u.RawQuery = q.Encode()
	}
	return u.String(), true
}

func writeCallbackPage(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	title, body := "Signed in to codeforge", "You can close this tab and go back to your terminal."
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		title, body = "Sign-in failed", html.EscapeString(err.Error())
	}
	fmt.Fprintf(w, "<!doctype html><meta charset=utf-8><title>%s</title>"+
		"<body style=\"font-family:system-ui;margin:4em auto;max-width:32em\"><h1>%s</h1><p>%s</p>",
		title, title, body)
}

// Device signs in with a code shown in the terminal.
func (f Flow) Device(ctx context.Context, cfg *CLIConfig, d *Discovery) (*oauth2.Token, error) {
	if d.DeviceAuthorizationEndpoint == "" {
		return nil, ErrDeviceUnavailable
	}
	oc := f.oauthConfig(cfg, d, "")
	ctx = f.withClient(ctx)
	auth, err := oc.DeviceAuth(ctx)
	if err != nil {
		if isRefusedGrant(err) {
			return nil, ErrDeviceUnavailable
		}
		return nil, fmt.Errorf("starting the device sign-in: %w", describe(err))
	}
	fmt.Fprintf(f.Out, "To sign in, open %s and enter the code:\n\n  %s\n\n", auth.VerificationURI, auth.UserCode)
	if f.OpenBrowser != nil && auth.VerificationURIComplete != "" {
		_ = f.OpenBrowser(auth.VerificationURIComplete)
	}
	token, err := oc.DeviceAccessToken(ctx, auth)
	if err != nil {
		if isRefusedGrant(err) {
			return nil, ErrDeviceUnavailable
		}
		return nil, fmt.Errorf("device sign-in: %w", describe(err))
	}
	return token, nil
}

// Refresh exchanges a refresh token for new tokens.
func Refresh(ctx context.Context, client *http.Client, cfg *CLIConfig, d *Discovery, refreshToken string) (*oauth2.Token, error) {
	oc := (Flow{HTTP: client}).oauthConfig(cfg, d, "")
	ctx = context.WithValue(ctx, oauth2.HTTPClient, client)
	expired := &oauth2.Token{RefreshToken: refreshToken, Expiry: time.Unix(1, 0)}
	token, err := oc.TokenSource(ctx, expired).Token()
	if err != nil {
		return nil, describe(err)
	}
	if token.RefreshToken == "" {
		token.RefreshToken = refreshToken
	}
	return token, nil
}

// Revoke revokes a token (RFC 7009). Revoking an unknown token succeeds.
func Revoke(ctx context.Context, client *http.Client, endpoint, clientID, token, hint string) error {
	form := url.Values{"token": {token}, "token_type_hint": {hint}, "client_id": {clientID}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("revocation answered %s", resp.Status)
	}
	return nil
}

func isRefusedGrant(err error) bool {
	var re *oauth2.RetrieveError
	if !errors.As(err, &re) {
		return false
	}
	switch re.ErrorCode {
	case "unauthorized_client", "unsupported_grant_type", "invalid_client":
		return true
	}
	return false
}

// describe keeps the OAuth error code and description, not the raw body.
func describe(err error) error {
	var re *oauth2.RetrieveError
	if errors.As(err, &re) && re.ErrorCode != "" {
		if re.ErrorDescription != "" {
			return fmt.Errorf("%s: %s", re.ErrorCode, re.ErrorDescription)
		}
		return errors.New(re.ErrorCode)
	}
	return err
}

func getJSON(ctx context.Context, client *http.Client, endpoint string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s answered %s", endpoint, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func randomString() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
