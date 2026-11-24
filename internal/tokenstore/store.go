// Package tokenstore keeps the CLI's sign-ins, one per host.
//
// Tokens go to the operating system's keychain (macOS Keychain, Windows
// Credential Manager, the Secret Service on Linux). When no keychain is
// available — a server, a container, WSL without a desktop session — they go
// to a tokens.json file readable by the user only (0600) in the configuration
// directory. PJ_TOKEN_STORE=file forces the file.
package tokenstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/zalando/go-keyring"
)

// ErrNotFound means there is no sign-in for the host.
var ErrNotFound = errors.New("not signed in")

// Token is one sign-in: the OIDC tokens and where they are valid.
type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
	Expiry       time.Time `json:"expiry,omitempty"`
	IDToken      string    `json:"id_token,omitempty"`
	// ClientID and Issuer are needed to refresh and revoke the tokens.
	ClientID string `json:"client_id"`
	Issuer   string `json:"issuer"`
	// APIURL is the API the access token is for (from /api/cli/config).
	APIURL string `json:"api_url"`
}

// Store reads and writes sign-ins by host.
type Store interface {
	Get(host string) (*Token, error)
	Set(host string, token *Token) error
	Delete(host string) error
}

const keyringService = "pointjungle"

// Open returns the store to use: the keychain with a file fallback, or the
// file only when PJ_TOKEN_STORE=file.
func Open(dir string, getenv func(string) string) Store {
	file := &FileStore{Path: filepath.Join(dir, "tokens.json")}
	if getenv("PJ_TOKEN_STORE") == "file" {
		return file
	}
	return &autoStore{keyring: KeyringStore{}, file: file}
}

// KeyringStore keeps tokens in the OS keychain, one entry per host.
type KeyringStore struct{}

func (KeyringStore) Get(host string) (*Token, error) {
	value, err := keyring.Get(keyringService, host)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var token Token
	if err := json.Unmarshal([]byte(value), &token); err != nil {
		return nil, fmt.Errorf("keychain entry for %s: %w", host, err)
	}
	return &token, nil
}

func (KeyringStore) Set(host string, token *Token) error {
	value, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return keyring.Set(keyringService, host, string(value))
}

func (KeyringStore) Delete(host string) error {
	err := keyring.Delete(keyringService, host)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}

// FileStore keeps all tokens in one JSON file, mode 0600, directory 0700.
type FileStore struct {
	Path string
}

func (s *FileStore) read() (map[string]*Token, error) {
	data, err := os.ReadFile(s.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]*Token{}, nil
	}
	if err != nil {
		return nil, err
	}
	tokens := map[string]*Token{}
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("%s: %w", s.Path, err)
	}
	return tokens, nil
}

func (s *FileStore) write(tokens map[string]*Token) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.Path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.Path)
}

func (s *FileStore) Get(host string) (*Token, error) {
	tokens, err := s.read()
	if err != nil {
		return nil, err
	}
	token, ok := tokens[host]
	if !ok {
		return nil, ErrNotFound
	}
	return token, nil
}

func (s *FileStore) Set(host string, token *Token) error {
	tokens, err := s.read()
	if err != nil {
		return err
	}
	tokens[host] = token
	return s.write(tokens)
}

func (s *FileStore) Delete(host string) error {
	tokens, err := s.read()
	if err != nil {
		return err
	}
	if _, ok := tokens[host]; !ok {
		return nil
	}
	delete(tokens, host)
	return s.write(tokens)
}

// autoStore uses the keychain and falls back to the file when the keychain
// does not work. Reads look in both, so a sign-in made without a keychain is
// still found once one is available.
type autoStore struct {
	keyring Store
	file    Store
}

func (s *autoStore) Get(host string) (*Token, error) {
	token, err := s.keyring.Get(host)
	if err == nil {
		return token, nil
	}
	return s.file.Get(host)
}

func (s *autoStore) Set(host string, token *Token) error {
	if err := s.keyring.Set(host, token); err == nil {
		_ = s.file.Delete(host)
		return nil
	}
	return s.file.Set(host, token)
}

func (s *autoStore) Delete(host string) error {
	keyringErr := s.keyring.Delete(host)
	fileErr := s.file.Delete(host)
	if fileErr != nil {
		return fileErr
	}
	_ = keyringErr // no keychain: nothing was stored there
	return nil
}
