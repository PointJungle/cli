package cli

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/pointjungle/codeforge-cli/internal/api"
	"github.com/pointjungle/codeforge-cli/internal/signin"
	"github.com/pointjungle/codeforge-cli/internal/tokenstore"
)

type userResult struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name"`
}

type statusResult struct {
	Host      string      `json:"host"`
	SignedIn  bool        `json:"signed_in"`
	User      *userResult `json:"user,omitempty"`
	ExpiresAt *time.Time  `json:"expires_at,omitempty"`
}

func (a *App) authCommand() *cobra.Command {
	auth := &cobra.Command{
		Use:   "auth",
		Short: "Sign in to codeforge, sign out, see who you are",
		Args:  noArgs,
	}
	auth.AddCommand(a.loginCommand(), a.logoutCommand(), a.statusCommand())
	return auth
}

func (a *App) loginCommand() *cobra.Command {
	var device, noBrowser bool
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Sign in through codeforge's identity provider",
		Long: "Opens your browser on codeforge's sign-in page and waits for it to\n" +
			"redirect to a local port. With --device, shows a code to enter on any\n" +
			"device instead. One sign-in is kept per host.",
		Args: noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := a.settings()
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), a.LoginTimeout)
			defer cancel()

			cfg, err := signin.FetchCLIConfig(ctx, a.HTTP, s.Host)
			if err != nil {
				return err
			}
			discovery, err := signin.Discover(ctx, a.HTTP, cfg.Issuer)
			if err != nil {
				return err
			}
			flow := signin.Flow{HTTP: a.HTTP, OpenBrowser: a.OpenBrowser, Out: a.Stderr}
			if noBrowser {
				flow.OpenBrowser = nil
			}
			login := flow.Loopback
			if device {
				login = flow.Device
			}
			token, err := login(ctx, cfg, discovery)
			if errors.Is(err, signin.ErrDeviceUnavailable) {
				return &cliError{exit: ExitError, code: "device_unavailable",
					msg: fmt.Sprintf("device sign-in is not available on %s yet; run codeforge auth login without --device", s.Host)}
			}
			if err != nil {
				return err
			}
			stored := &tokenstore.Token{
				AccessToken:  token.AccessToken,
				RefreshToken: token.RefreshToken,
				TokenType:    token.Type(),
				Expiry:       token.Expiry,
				ClientID:     cfg.ClientID,
				Issuer:       cfg.Issuer,
				APIURL:       cfg.APIURL,
			}
			if id, ok := token.Extra("id_token").(string); ok {
				stored.IDToken = id
			}
			if err := a.store(s).Set(s.Host, stored); err != nil {
				return fmt.Errorf("saving the sign-in: %w", err)
			}
			user, err := a.apiClient(s).Me(ctx)
			if err != nil {
				return apiError(err, s.Host)
			}
			a.result(statusResult{Host: s.Host, SignedIn: true, User: toUser(user), ExpiresAt: expiry(stored.Expiry)},
				fmt.Sprintf("Signed in to %s as %s.", s.Host, user.Email))
			return nil
		},
	}
	cmd.Flags().BoolVar(&device, "device", false, "sign in with a code to enter on another device")
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "print the sign-in URL without opening a browser")
	return cmd
}

func (a *App) statusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show who you are signed in as on the host",
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := a.settings()
			if err != nil {
				return err
			}
			client := a.apiClient(s)
			user, err := client.Me(cmd.Context())
			if err != nil {
				return apiError(err, s.Host)
			}
			token, err := client.Token(cmd.Context())
			if err != nil {
				return apiError(err, s.Host)
			}
			a.result(statusResult{Host: s.Host, SignedIn: true, User: toUser(user), ExpiresAt: expiry(token.Expiry)},
				fmt.Sprintf("Signed in to %s as %s.", s.Host, user.Email))
			return nil
		},
	}
}

func (a *App) logoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Sign out of the host and forget its tokens",
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := a.settings()
			if err != nil {
				return err
			}
			store := a.store(s)
			token, err := store.Get(s.Host)
			if err != nil && !errors.Is(err, tokenstore.ErrNotFound) {
				return err
			}
			if token != nil {
				a.revoke(cmd.Context(), token)
				if err := store.Delete(s.Host); err != nil {
					return err
				}
			}
			a.result(map[string]any{"host": s.Host, "signed_out": true}, fmt.Sprintf("Signed out of %s.", s.Host))
			return nil
		},
	}
}

// revoke revokes the refresh token, best effort: the local sign-in is removed
// even when the identity provider cannot be reached.
func (a *App) revoke(ctx context.Context, token *tokenstore.Token) {
	if token.RefreshToken == "" || token.Issuer == "" {
		return
	}
	discovery, err := signin.Discover(ctx, a.HTTP, token.Issuer)
	if err == nil && discovery.RevocationEndpoint != "" {
		err = signin.Revoke(ctx, a.HTTP, discovery.RevocationEndpoint, token.ClientID, token.RefreshToken, "refresh_token")
	}
	if err != nil {
		fmt.Fprintf(a.Stderr, "warning: could not revoke the session on the server: %v\n", err)
	}
}

func toUser(u *api.User) *userResult {
	return &userResult{ID: u.ID, Email: u.Email, DisplayName: u.DisplayName}
}

func expiry(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
