package tokenstore

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestFileStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "tokens.json")
	s := &FileStore{Path: path}

	if _, err := s.Get("https://a"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("empty store: %v", err)
	}
	if err := s.Set("https://a", &Token{AccessToken: "x"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Set("https://b", &Token{AccessToken: "y"}); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get("https://a")
	if err != nil || got.AccessToken != "x" {
		t.Fatalf("get: %v %v", got, err)
	}
	info, _ := os.Stat(path)
	dirInfo, _ := os.Stat(filepath.Dir(path))
	if info.Mode().Perm() != 0o600 || dirInfo.Mode().Perm() != 0o700 {
		t.Fatalf("modes: file %v, dir %v", info.Mode(), dirInfo.Mode())
	}
	if err := s.Delete("https://a"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get("https://a"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("after delete: %v", err)
	}
	if got, _ := s.Get("https://b"); got == nil || got.AccessToken != "y" {
		t.Fatalf("other host must be kept: %v", got)
	}
}

func TestOpenUsesKeychainWhenAvailable(t *testing.T) {
	keyring.MockInit()
	dir := t.TempDir()
	s := Open(dir, func(string) string { return "" })
	if err := s.Set("https://a", &Token{AccessToken: "x"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "tokens.json")); !os.IsNotExist(err) {
		t.Fatal("with a keychain, no tokens file must be written")
	}
	got, err := s.Get("https://a")
	if err != nil || got.AccessToken != "x" {
		t.Fatalf("get: %v %v", got, err)
	}
	if err := s.Delete("https://a"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get("https://a"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("after delete: %v", err)
	}
}

func TestOpenFallsBackToFileWithoutKeychain(t *testing.T) {
	keyring.MockInitWithError(errors.New("no secret service"))
	dir := t.TempDir()
	s := Open(dir, func(string) string { return "" })
	if err := s.Set("https://a", &Token{AccessToken: "x"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "tokens.json")); err != nil {
		t.Fatalf("without a keychain, tokens must go to the file: %v", err)
	}
	if got, err := s.Get("https://a"); err != nil || got.AccessToken != "x" {
		t.Fatalf("get: %v %v", got, err)
	}
}

func TestFileStoreForced(t *testing.T) {
	s := Open(t.TempDir(), func(k string) string {
		if k == "PJ_TOKEN_STORE" {
			return "file"
		}
		return ""
	})
	if _, ok := s.(*FileStore); !ok {
		t.Fatalf("PJ_TOKEN_STORE=file must give the file store, got %T", s)
	}
}
