package localserver

import (
	"path/filepath"
	"testing"
)

func TestParseEntries(t *testing.T) {
	root := t.TempDir()
	appPath := filepath.Join(root, "app")
	apiPath := filepath.Join(root, "api")

	opts := EntryParseOptions{
		EntryName:     "test entry",
		AliasName:     "test alias",
		PortName:      "test port",
		DuplicateName: "test alias",
	}

	t.Run("path only", func(t *testing.T) {
		entry, err := ParseEntry(appPath, opts)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if entry.Path != appPath {
			t.Fatalf("expected path %q, got %q", appPath, entry.Path)
		}

		if entry.Alias != "app" {
			t.Fatalf("expected alias app, got %q", entry.Alias)
		}
	})

	t.Run("alias and port", func(t *testing.T) {
		entry, err := ParseEntry(appPath+"@frontend:9090", opts)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if entry.Alias != "frontend" {
			t.Fatalf("expected alias frontend, got %q", entry.Alias)
		}

		if entry.Port != 9090 {
			t.Fatalf("expected port 9090, got %d", entry.Port)
		}
	})

	t.Run("duplicates", func(t *testing.T) {
		_, err := ParseEntries([]string{
			appPath + "@api",
			apiPath + "@api",
		}, opts)
		if err == nil {
			t.Fatal("expected duplicate error, got nil")
		}

		if err.Error() != `duplicate test alias "api"` {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("legacy port before alias", func(t *testing.T) {
		_, err := ParseEntry(appPath+":9090@app", opts)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != `invalid test entry "`+appPath+`:9090@app": use <path>@<alias>:<port>` {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
