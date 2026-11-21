// Package api calls the codeforge API as the signed-in person.
//
// Each request carries the stored access token (Authorization: Bearer). When the
// token has expired — or the API answers 401 — the client refreshes it once with
// the refresh token, stores the new tokens and retries. If that fails, the
// person has to sign in again (ErrNotSignedIn).
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pointjungle/codeforge-cli/internal/signin"
	"github.com/pointjungle/codeforge-cli/internal/tokenstore"
)

var (
	// ErrNotSignedIn: no sign-in for the host, or it can no longer be refreshed.
	ErrNotSignedIn = errors.New("not signed in")
	// ErrDenied: the API refused the request (403) or found nothing (404).
	ErrDenied = errors.New("permission denied or not found")
)

// expiryMargin refreshes tokens slightly before they expire.
const expiryMargin = 30 * time.Second

// Client calls the API of one host.
type Client struct {
	HTTP  *http.Client
	Host  string
	Store tokenstore.Store
	Now   func() time.Time
}

// User is the signed-in person (GET /api/profiles/me/).
type User struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name"`
}

// Me returns the signed-in person.
func (c *Client) Me(ctx context.Context) (*User, error) {
	var body struct {
		User User `json:"user"`
	}
	if err := c.GetJSON(ctx, "/api/profiles/me/", &body); err != nil {
		return nil, err
	}
	return &body.User, nil
}

// Token returns the stored sign-in, refreshed if it has expired.
func (c *Client) Token(ctx context.Context) (*tokenstore.Token, error) {
	token, err := c.Store.Get(c.Host)
	if errors.Is(err, tokenstore.ErrNotFound) {
		return nil, ErrNotSignedIn
	}
	if err != nil {
		return nil, err
	}
	if !token.Expiry.IsZero() && c.now().Add(expiryMargin).After(token.Expiry) {
		return c.refresh(ctx, token)
	}
	return token, nil
}

// GetJSON GETs an API path and decodes the JSON answer into out.
func (c *Client) GetJSON(ctx context.Context, path string, out any) error {
	resp, err := c.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}

// Do sends an authenticated request; a 2xx response is returned open.
func (c *Client) Do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	token, err := c.Token(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := c.send(ctx, token, method, path, body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized && body == nil {
		resp.Body.Close()
		if token, err = c.refresh(ctx, token); err != nil {
			return nil, err
		}
		if resp, err = c.send(ctx, token, method, path, nil); err != nil {
			return nil, err
		}
	}
	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		resp.Body.Close()
		return nil, ErrNotSignedIn
	case resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound:
		resp.Body.Close()
		return nil, ErrDenied
	case resp.StatusCode >= 300:
		defer resp.Body.Close()
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("%s %s answered %s: %s", method, path, resp.Status, strings.TrimSpace(string(detail)))
	}
	return resp, nil
}

func (c *Client) send(ctx context.Context, token *tokenstore.Token, method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(token.APIURL, "/")+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.HTTP.Do(req)
}

func (c *Client) refresh(ctx context.Context, token *tokenstore.Token) (*tokenstore.Token, error) {
	if token.RefreshToken == "" {
		return nil, ErrNotSignedIn
	}
	discovery, err := signin.Discover(ctx, c.HTTP, token.Issuer)
	if err != nil {
		return nil, err
	}
	cfg := &signin.CLIConfig{ClientID: token.ClientID, Issuer: token.Issuer, APIURL: token.APIURL}
	fresh, err := signin.Refresh(ctx, c.HTTP, cfg, discovery, token.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("%w (%v)", ErrNotSignedIn, err)
	}
	updated := *token
	updated.AccessToken = fresh.AccessToken
	updated.RefreshToken = fresh.RefreshToken
	updated.Expiry = fresh.Expiry
	if t := fresh.Type(); t != "" {
		updated.TokenType = t
	}
	if id, ok := fresh.Extra("id_token").(string); ok && id != "" {
		updated.IDToken = id
	}
	if err := c.Store.Set(c.Host, &updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

func (c *Client) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}
