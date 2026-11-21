// Package cli defines the codeforge commands.
//
// Output contract — scripts and AI agents rely on it:
//
//   - Human output by default. With --json, every command writes exactly one
//     JSON document to stdout — its result, or {"error": {"code", "message"}} —
//     and messages for people go to stderr.
//   - Stable exit codes: 0 success, 1 error, 2 usage error, 3 not signed in,
//     4 permission denied or not found.
package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/pointjungle/codeforge-cli/internal/api"
	"github.com/pointjungle/codeforge-cli/internal/browser"
	"github.com/pointjungle/codeforge-cli/internal/config"
	"github.com/pointjungle/codeforge-cli/internal/httpclient"
	"github.com/pointjungle/codeforge-cli/internal/tokenstore"
)

// Exit codes.
const (
	ExitOK          = 0
	ExitError       = 1
	ExitUsage       = 2
	ExitNotSignedIn = 3
	ExitDenied      = 4
)

// Build information, set at build time.
type Build struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

// App holds what commands need from the outside world, so that tests can
// replace it.
type App struct {
	Build       Build
	Stdout      io.Writer
	Stderr      io.Writer
	Getenv      func(string) string
	Cwd         string
	HTTP        *http.Client
	OpenBrowser func(string) error
	Now         func() time.Time
	// LoginTimeout bounds how long a sign-in waits for the person.
	LoginTimeout time.Duration

	// set from the global flags
	host string
	json bool
}

// NewApp returns the App of a real run.
func NewApp(build Build) *App {
	cwd, _ := os.Getwd()
	return &App{
		Build:        build,
		Stdout:       os.Stdout,
		Stderr:       os.Stderr,
		Getenv:       os.Getenv,
		Cwd:          cwd,
		HTTP:         httpclient.New(30*time.Second, httpclient.UserAgent(build.Version)),
		OpenBrowser:  browser.Open,
		Now:          time.Now,
		LoginTimeout: 5 * time.Minute,
	}
}

// cliError is an error with its exit code and a stable machine code.
type cliError struct {
	exit int
	code string
	msg  string
}

func (e *cliError) Error() string { return e.msg }

func usageError(err error) error {
	return &cliError{exit: ExitUsage, code: "usage", msg: err.Error()}
}

func notSignedIn(host string) error {
	login := "codeforge auth login"
	if host != config.DefaultHost {
		login += " --host " + host
	}
	return &cliError{
		exit: ExitNotSignedIn,
		code: "not_signed_in",
		msg:  fmt.Sprintf("not signed in to %s; run: %s", host, login),
	}
}

// Run executes the command line and returns the exit code.
func (a *App) Run(args []string) int {
	root := a.rootCommand()
	root.SetArgs(args)
	root.SetOut(a.Stdout)
	root.SetErr(a.Stderr)
	err := root.Execute()
	if err == nil {
		return ExitOK
	}
	var ce *cliError
	if !errors.As(err, &ce) {
		ce = &cliError{exit: ExitError, code: "error", msg: err.Error()}
		if isCobraUsageError(err) {
			ce = &cliError{exit: ExitUsage, code: "usage", msg: err.Error()}
		}
	}
	// When the command line does not parse, --json may not have been read yet.
	if a.json || wantsJSON(args) {
		a.writeJSON(map[string]any{"error": map[string]string{"code": ce.code, "message": ce.msg}})
	} else {
		fmt.Fprintf(a.Stderr, "error: %s\n", ce.msg)
	}
	return ce.exit
}

func wantsJSON(args []string) bool {
	for _, arg := range args {
		if arg == "--" {
			return false
		}
		if arg == "--json" || arg == "--json=true" {
			return true
		}
	}
	return false
}

func isCobraUsageError(err error) bool {
	msg := err.Error()
	for _, prefix := range []string{"unknown command", "unknown flag", "unknown shorthand flag", "accepts ", "requires ", "invalid argument", "flag needs an argument"} {
		if strings.HasPrefix(msg, prefix) {
			return true
		}
	}
	return false
}

func (a *App) rootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "codeforge",
		Short:         "Work with codeforge from your terminal",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	root.PersistentFlags().StringVar(&a.host, "host", "", "codeforge API host (default: from CODEFORGE_HOST, .codeforge/, config.toml, or production)")
	root.PersistentFlags().BoolVar(&a.json, "json", false, "write one JSON document to stdout")
	// --host targets another codeforge (staging, a local backend): it is meant
	// for codeforge's developers, so it stays out of --help.
	_ = root.PersistentFlags().MarkHidden("host")
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return usageError(err) })
	root.AddCommand(a.authCommand(), a.versionCommand())
	return root
}

func noArgs(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return usageError(fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath()))
	}
	return nil
}

// settings resolves the settings of the current command.
func (a *App) settings() (config.Settings, error) {
	return config.Resolve(a.host, a.Cwd, a.Getenv)
}

func (a *App) store(s config.Settings) tokenstore.Store {
	return tokenstore.Open(s.Dir, a.Getenv)
}

func (a *App) apiClient(s config.Settings) *api.Client {
	return &api.Client{HTTP: a.HTTP, Host: s.Host, Store: a.store(s), Now: a.Now}
}

// result writes a command's result: JSON with --json, the human text otherwise.
func (a *App) result(value any, human string) {
	if a.json {
		a.writeJSON(value)
		return
	}
	fmt.Fprintln(a.Stdout, human)
}

func (a *App) writeJSON(value any) {
	enc := json.NewEncoder(a.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(value)
}

// apiError maps API client errors to exit codes.
func apiError(err error, host string) error {
	switch {
	case errors.Is(err, api.ErrNotSignedIn):
		return notSignedIn(host)
	case errors.Is(err, api.ErrDenied):
		return &cliError{exit: ExitDenied, code: "denied", msg: err.Error()}
	}
	return err
}
