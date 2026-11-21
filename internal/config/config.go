// Package config resolves the CLI's settings.
//
// Every setting is taken from the first place that defines it:
//
//  1. a command-line flag (--host);
//  2. an environment variable (CODEFORGE_HOST);
//  3. the directory's link, .codeforge/project.json in the current directory
//     or one of its parents (written by `codeforge link`);
//  4. the user's configuration file, config.toml in the configuration
//     directory;
//  5. the built-in default.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// DefaultHost is the production API.
const DefaultHost = "https://api.codeforge.pointjungle.com"

// Source tells where a setting came from.
type Source string

const (
	SourceFlag    Source = "flag"
	SourceEnv     Source = "env"
	SourceProject Source = "project"
	SourceConfig  Source = "config"
	SourceDefault Source = "default"
)

// Settings are the resolved settings of one command.
type Settings struct {
	// Host is the base URL of the codeforge API, e.g. https://api.codeforge.pointjungle.com.
	Host       string `json:"host"`
	HostSource Source `json:"host_source"`
	// Dir is the configuration directory (config.toml, tokens file).
	Dir string `json:"config_dir"`
}

type fileConfig struct {
	Host string `toml:"host"`
}

type projectLink struct {
	Host string `json:"host"`
}

// Dir returns the configuration directory: $CODEFORGE_CONFIG_DIR, or
// "codeforge" in the OS's user configuration directory (~/.config on Linux).
func Dir(getenv func(string) string) (string, error) {
	if dir := getenv("CODEFORGE_CONFIG_DIR"); dir != "" {
		return dir, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("no configuration directory: %w", err)
	}
	return filepath.Join(base, "codeforge"), nil
}

// Resolve resolves the settings for a command run in cwd.
func Resolve(flagHost, cwd string, getenv func(string) string) (Settings, error) {
	dir, err := Dir(getenv)
	if err != nil {
		return Settings{}, err
	}
	raw, source := DefaultHost, SourceDefault
	switch {
	case flagHost != "":
		raw, source = flagHost, SourceFlag
	case getenv("CODEFORGE_HOST") != "":
		raw, source = getenv("CODEFORGE_HOST"), SourceEnv
	default:
		if host, err := projectHost(cwd); err != nil {
			return Settings{}, err
		} else if host != "" {
			raw, source = host, SourceProject
		} else if host, err := configHost(dir); err != nil {
			return Settings{}, err
		} else if host != "" {
			raw, source = host, SourceConfig
		}
	}
	host, err := NormalizeHost(raw)
	if err != nil {
		return Settings{}, fmt.Errorf("host from %s: %w", source, err)
	}
	return Settings{Host: host, HostSource: source, Dir: dir}, nil
}

// NormalizeHost turns what a person types into the API's base URL:
// "staging.example.com" → "https://staging.example.com",
// "localhost:8000" → "http://localhost:8000" (loopback hosts, *.localhost
// included, default to http).
// Paths, queries and trailing slashes are not allowed.
func NormalizeHost(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("empty host")
	}
	if !strings.Contains(raw, "://") {
		scheme := "https"
		if isLoopback(raw) {
			scheme = "http"
		}
		raw = scheme + "://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("invalid host %q", raw)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("invalid host %q: scheme must be https or http", raw)
	}
	if (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" {
		return "", fmt.Errorf("invalid host %q: give the host only, without a path", raw)
	}
	return u.Scheme + "://" + u.Host, nil
}

func isLoopback(hostport string) bool {
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	}
	host = strings.Trim(host, "[]")
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	return host == "localhost" || strings.HasSuffix(host, ".localhost") || host == "127.0.0.1" || host == "::1"
}

// projectHost reads the host of the nearest .codeforge/project.json.
func projectHost(cwd string) (string, error) {
	for dir := cwd; ; dir = filepath.Dir(dir) {
		path := filepath.Join(dir, ".codeforge", "project.json")
		data, err := os.ReadFile(path)
		if err == nil {
			var link projectLink
			if err := json.Unmarshal(data, &link); err != nil {
				return "", fmt.Errorf("%s: %w", path, err)
			}
			return link.Host, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		if parent := filepath.Dir(dir); parent == dir {
			return "", nil
		}
	}
}

func configHost(dir string) (string, error) {
	path := filepath.Join(dir, "config.toml")
	var cfg fileConfig
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", nil
		}
		return "", fmt.Errorf("%s: %w", path, err)
	}
	return cfg.Host, nil
}
