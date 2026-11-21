package config

import (
	"os"
	"path/filepath"
	"testing"
)

func env(values map[string]string) func(string) string {
	return func(k string) string { return values[k] }
}

func TestNormalizeHost(t *testing.T) {
	cases := map[string]string{
		"api.codeforge.pointjungle.com":          "https://api.codeforge.pointjungle.com",
		"https://api.codeforge.pointjungle.com/": "https://api.codeforge.pointjungle.com",
		"localhost:8000":                         "http://localhost:8000",
		"api.localhost:8000":                     "http://api.localhost:8000",
		"127.0.0.1:8000":                         "http://127.0.0.1:8000",
		"[::1]:8000":                             "http://[::1]:8000",
		"http://staging.example.com":             "http://staging.example.com",
	}
	for in, want := range cases {
		got, err := NormalizeHost(in)
		if err != nil || got != want {
			t.Errorf("NormalizeHost(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, bad := range []string{"", "ftp://x", "https://x/api", "https://x?y=1"} {
		if _, err := NormalizeHost(bad); err == nil {
			t.Errorf("NormalizeHost(%q) should fail", bad)
		}
	}
}

func TestResolveOrder(t *testing.T) {
	dir := t.TempDir()
	project := t.TempDir()
	nested := filepath.Join(project, "a", "b")
	if err := os.MkdirAll(filepath.Join(project, ".codeforge"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(path, content string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(dir, "config.toml"), `host = "config.example.com"`)
	write(filepath.Join(project, ".codeforge", "project.json"), `{"host": "project.example.com"}`)

	base := map[string]string{"CODEFORGE_CONFIG_DIR": dir}
	withEnv := map[string]string{"CODEFORGE_CONFIG_DIR": dir, "CODEFORGE_HOST": "env.example.com"}

	cases := []struct {
		name   string
		flag   string
		cwd    string
		env    map[string]string
		host   string
		source Source
	}{
		{"flag wins", "flag.example.com", nested, withEnv, "https://flag.example.com", SourceFlag},
		{"then the environment", "", nested, withEnv, "https://env.example.com", SourceEnv},
		{"then the directory's link, found in a parent", "", nested, base, "https://project.example.com", SourceProject},
		{"then config.toml", "", t.TempDir(), base, "https://config.example.com", SourceConfig},
		{"then the default", "", t.TempDir(), map[string]string{"CODEFORGE_CONFIG_DIR": t.TempDir()}, DefaultHost, SourceDefault},
	}
	for _, c := range cases {
		got, err := Resolve(c.flag, c.cwd, env(c.env))
		if err != nil || got.Host != c.host || got.HostSource != c.source {
			t.Errorf("%s: got %+v, %v; want %s from %s", c.name, got, err, c.host, c.source)
		}
	}
}

func TestResolveRejectsBrokenFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte("host = "), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve("", t.TempDir(), env(map[string]string{"CODEFORGE_CONFIG_DIR": dir})); err == nil {
		t.Fatal("a broken config.toml must be reported")
	}
}
