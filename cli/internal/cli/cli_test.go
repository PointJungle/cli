package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pointjungle/codeforge-sdk/cli/internal/config"
)

type run struct {
	code   int
	stdout string
	stderr string
}

func (r run) json(t *testing.T) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal([]byte(r.stdout), &doc); err != nil {
		t.Fatalf("stdout is not one JSON document: %v\n%s", err, r.stdout)
	}
	return doc
}

type harness struct {
	t      *testing.T
	fake   *fakeCodeforge
	dir    string
	env    map[string]string
	now    func() time.Time
	browse func(string) error
}

func newHarness(t *testing.T) *harness {
	fake := newFake(t)
	dir := t.TempDir()
	return &harness{
		t:    t,
		fake: fake,
		dir:  dir,
		env: map[string]string{
			"CODEFORGE_CONFIG_DIR":  dir,
			"CODEFORGE_TOKEN_STORE": "file",
			"CODEFORGE_HOST":        fake.URL(),
		},
		now:    time.Now,
		browse: browse(t),
	}
}

func (h *harness) run(args ...string) run {
	var stdout, stderr bytes.Buffer
	app := &App{
		Build:        Build{Version: "1.2.3", Commit: "abc", Date: "2026-11-20"},
		Stdout:       &stdout,
		Stderr:       &stderr,
		Getenv:       func(k string) string { return h.env[k] },
		Cwd:          h.t.TempDir(),
		HTTP:         h.fake.server.Client(),
		OpenBrowser:  h.browse,
		Now:          h.now,
		LoginTimeout: 10 * time.Second,
	}
	code := app.Run(args)
	return run{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

func (h *harness) storedTokens() map[string]map[string]any {
	data, err := os.ReadFile(filepath.Join(h.dir, "tokens.json"))
	if err != nil {
		h.t.Fatalf("tokens file: %v", err)
	}
	tokens := map[string]map[string]any{}
	_ = json.Unmarshal(data, &tokens)
	return tokens
}

// --- sign-in -----------------------------------------------------------------

func TestLoginStatusLogout(t *testing.T) {
	h := newHarness(t)

	r := h.run("auth", "login", "--json")
	if r.code != ExitOK {
		t.Fatalf("login exit %d: %s %s", r.code, r.stdout, r.stderr)
	}
	doc := r.json(t)
	user := doc["user"].(map[string]any)
	if doc["host"] != h.fake.URL() || user["email"] != "ada@example.com" || doc["signed_in"] != true {
		t.Fatalf("unexpected login result: %v", doc)
	}
	if !strings.Contains(r.stderr, "/authorize?") {
		t.Errorf("the sign-in URL should be printed for the person: %q", r.stderr)
	}

	info, err := os.Stat(filepath.Join(h.dir, "tokens.json"))
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("tokens file must exist with mode 0600: %v %v", info, err)
	}
	stored := h.storedTokens()[h.fake.URL()]
	if stored["client_id"] != clientID || stored["refresh_token"] == "" || stored["api_url"] != h.fake.URL() {
		t.Fatalf("unexpected stored sign-in: %v", stored)
	}

	r = h.run("auth", "status", "--json")
	if r.code != ExitOK || r.json(t)["user"].(map[string]any)["email"] != "ada@example.com" {
		t.Fatalf("status: %d %s", r.code, r.stdout)
	}

	refresh := stored["refresh_token"].(string)
	r = h.run("auth", "logout", "--json")
	if r.code != ExitOK || r.json(t)["signed_out"] != true {
		t.Fatalf("logout: %d %s", r.code, r.stdout)
	}
	if len(h.fake.revoked) != 1 || h.fake.revoked[0] != refresh {
		t.Fatalf("logout must revoke the refresh token, revoked: %v", h.fake.revoked)
	}

	r = h.run("auth", "status", "--json")
	if r.code != ExitNotSignedIn || r.json(t)["error"].(map[string]any)["code"] != "not_signed_in" {
		t.Fatalf("status after logout: %d %s", r.code, r.stdout)
	}
}

func TestHumanOutput(t *testing.T) {
	h := newHarness(t)
	r := h.run("auth", "login")
	if r.code != ExitOK || strings.TrimSpace(r.stdout) != "Signed in to "+h.fake.URL()+" as ada@example.com." {
		t.Fatalf("login: %d %q", r.code, r.stdout)
	}
	r = h.run("auth", "status")
	if r.code != ExitOK || !strings.Contains(r.stdout, "ada@example.com") {
		t.Fatalf("status: %d %q", r.code, r.stdout)
	}
}

func TestLoginRefusedByThePerson(t *testing.T) {
	h := newHarness(t)
	h.fake.denyAuthorize = true
	r := h.run("auth", "login", "--json")
	if r.code != ExitError || !strings.Contains(r.json(t)["error"].(map[string]any)["message"].(string), "access_denied") {
		t.Fatalf("refused login: %d %s", r.code, r.stdout)
	}
}

func TestLoginWithoutBrowserTimesOut(t *testing.T) {
	h := newHarness(t)
	h.browse = func(string) error { return nil } // nobody completes the sign-in
	var stdout, stderr bytes.Buffer
	app := &App{
		Stdout: &stdout, Stderr: &stderr,
		Getenv: func(k string) string { return h.env[k] }, Cwd: t.TempDir(),
		HTTP: h.fake.server.Client(), OpenBrowser: h.browse, Now: time.Now,
		LoginTimeout: 200 * time.Millisecond,
	}
	if code := app.Run([]string{"auth", "login"}); code != ExitError {
		t.Fatalf("exit %d, want %d: %s", code, ExitError, stderr.String())
	}
}

// --- device ------------------------------------------------------------------

func TestDeviceUnavailable(t *testing.T) {
	h := newHarness(t)
	r := h.run("auth", "login", "--device", "--json")
	if r.code != ExitError || r.json(t)["error"].(map[string]any)["code"] != "device_unavailable" {
		t.Fatalf("device login: %d %s", r.code, r.stdout)
	}
}

func TestDeviceLogin(t *testing.T) {
	h := newHarness(t)
	h.fake.deviceAllowed = true
	r := h.run("auth", "login", "--device", "--json")
	if r.code != ExitOK || r.json(t)["user"].(map[string]any)["email"] != "ada@example.com" {
		t.Fatalf("device login: %d %s %s", r.code, r.stdout, r.stderr)
	}
	if !strings.Contains(r.stderr, "ABCD-EFGH") {
		t.Errorf("the user code should be shown: %q", r.stderr)
	}
}

// --- refresh -----------------------------------------------------------------

func TestExpiredTokenIsRefreshed(t *testing.T) {
	h := newHarness(t)
	h.fake.expiresIn = 10 // within the 30 s refresh margin: already "expired"
	if r := h.run("auth", "login"); r.code != ExitOK {
		t.Fatalf("login: %s", r.stderr)
	}
	before := h.storedTokens()[h.fake.URL()]["access_token"]
	h.fake.expiresIn = 3600
	h.fake.refreshCalls = 0 // login's own /me call may already have refreshed

	r := h.run("auth", "status", "--json")
	if r.code != ExitOK {
		t.Fatalf("status: %d %s %s", r.code, r.stdout, r.stderr)
	}
	if h.fake.refreshCalls != 1 || h.storedTokens()[h.fake.URL()]["access_token"] == before {
		t.Fatalf("expected one refresh and a new stored access token (calls %d)", h.fake.refreshCalls)
	}
}

func TestRefusedTokenIsRefreshedOnce(t *testing.T) {
	h := newHarness(t)
	if r := h.run("auth", "login"); r.code != ExitOK {
		t.Fatalf("login: %s", r.stderr)
	}
	h.fake.invalidateAccessTokens() // e.g. revoked on the server
	r := h.run("auth", "status")
	if r.code != ExitOK || h.fake.refreshCalls != 1 {
		t.Fatalf("status: %d, refresh calls %d: %s", r.code, h.fake.refreshCalls, r.stderr)
	}
}

func TestFailedRefreshMeansNotSignedIn(t *testing.T) {
	h := newHarness(t)
	if r := h.run("auth", "login"); r.code != ExitOK {
		t.Fatalf("login: %s", r.stderr)
	}
	h.fake.invalidateAccessTokens()
	h.fake.refuseRefresh = true
	r := h.run("auth", "status", "--json")
	if r.code != ExitNotSignedIn {
		t.Fatalf("status: %d %s", r.code, r.stdout)
	}
}

// --- output contract ---------------------------------------------------------

func TestUsageErrors(t *testing.T) {
	h := newHarness(t)
	for _, args := range [][]string{{"nope"}, {"auth", "login", "--nope"}, {"auth", "status", "extra"}} {
		r := h.run(append(args, "--json")...)
		if r.code != ExitUsage || r.json(t)["error"].(map[string]any)["code"] != "usage" {
			t.Errorf("%v: exit %d %s", args, r.code, r.stdout)
		}
	}
}

func TestNotSignedInHumanMessage(t *testing.T) {
	h := newHarness(t)
	r := h.run("auth", "status")
	if r.code != ExitNotSignedIn || !strings.Contains(r.stderr, "codeforge auth login --host") {
		t.Fatalf("status: %d %q", r.code, r.stderr)
	}
}

func TestLogoutWhenSignedOutSucceeds(t *testing.T) {
	h := newHarness(t)
	if r := h.run("auth", "logout", "--json"); r.code != ExitOK {
		t.Fatalf("logout: %d %s", r.code, r.stdout)
	}
}

func TestVersion(t *testing.T) {
	h := newHarness(t)
	doc := h.run("version", "--json").json(t)
	if doc["version"] != "1.2.3" || doc["commit"] != "abc" {
		t.Fatalf("version: %v", doc)
	}
}

func TestHostFlagWinsOverEnvironment(t *testing.T) {
	h := newHarness(t)
	h.env["CODEFORGE_HOST"] = "unreachable.invalid"
	r := h.run("auth", "login", "--host", h.fake.URL(), "--json")
	if r.code != ExitOK || r.json(t)["host"] != h.fake.URL() {
		t.Fatalf("login with --host: %d %s %s", r.code, r.stdout, r.stderr)
	}
}

func TestNotSignedInHintOmitsTheDefaultHost(t *testing.T) {
	var ce *cliError
	if !errors.As(notSignedIn(config.DefaultHost), &ce) || ce.msg != "not signed in to "+config.DefaultHost+"; run: codeforge auth login" {
		t.Fatalf("default host: %v", ce)
	}
	if !errors.As(notSignedIn("http://localhost:8000"), &ce) || !strings.HasSuffix(ce.msg, "codeforge auth login --host http://localhost:8000") {
		t.Fatalf("other host: %v", ce)
	}
}

func TestHostFlagIsHiddenButWorks(t *testing.T) {
	h := newHarness(t)
	r := h.run("--help")
	if strings.Contains(r.stdout, "--host") {
		t.Fatalf("--host shown in help:\n%s", r.stdout)
	}
	delete(h.env, "CODEFORGE_HOST")
	if r := h.run("auth", "status", "--host", h.fake.URL()); r.code != ExitNotSignedIn || !strings.Contains(r.stderr, h.fake.URL()) {
		t.Fatalf("--host not accepted: %+v", r)
	}
}

func TestLoginShowsTheAccountChooserAndEndsOnTheSignedInPage(t *testing.T) {
	h := newHarness(t)
	if r := h.run("auth", "login"); r.code != ExitOK {
		t.Fatalf("login: %+v", r)
	}
	if len(h.fake.prompts) != 1 || h.fake.prompts[0] != "select_account" {
		t.Fatalf("prompt = %v, want select_account", h.fake.prompts)
	}
	if len(h.fake.landings) != 1 || h.fake.landings[0] != "" {
		t.Fatalf("signed-in page visits = %q, want one without error", h.fake.landings)
	}
}

func TestRefusedLoginEndsOnTheSignedInPageWithTheError(t *testing.T) {
	h := newHarness(t)
	h.fake.denyAuthorize = true
	if r := h.run("auth", "login"); r.code != ExitError {
		t.Fatalf("login: %+v", r)
	}
	if len(h.fake.landings) != 1 || h.fake.landings[0] != "error=access_denied" {
		t.Fatalf("signed-in page visits = %q, want error=access_denied", h.fake.landings)
	}
}

func TestLoginOnAnOlderHostShowsTheCLIPage(t *testing.T) {
	h := newHarness(t)
	h.fake.noSignedInURL = true
	var page string
	h.browse = func(authURL string) error {
		resp, err := http.Get(authURL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		page = string(body)
		return nil
	}
	if r := h.run("auth", "login"); r.code != ExitOK {
		t.Fatalf("login: %+v", r)
	}
	if !strings.Contains(page, "Signed in to codeforge") || len(h.fake.landings) != 0 {
		t.Fatalf("page = %q, landings = %q", page, h.fake.landings)
	}
}
