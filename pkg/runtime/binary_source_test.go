package runtime

import (
	"os"
	"path/filepath"
	"testing"

	ferretsource "github.com/MontFerret/ferret/v2/pkg/source"
)

func TestResolveBinaryScriptUsesMatchingLocalFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "query with spaces.fql")
	content := "RETURN 1"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write query: %v", err)
	}

	resolved, cleanup, err := resolveBinaryScript(ferretsource.New(path, content))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resolved != path {
		t.Fatalf("expected original path %q, got %q", path, resolved)
	}
	if err := cleanup(); err != nil {
		t.Fatalf("expected no cleanup error, got %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected original source to remain, got %v", err)
	}
}

func TestResolveBinaryScriptSnapshotsChangedLocalFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "query.fql")
	if err := os.WriteFile(path, []byte("RETURN 1"), 0o644); err != nil {
		t.Fatalf("failed to write query: %v", err)
	}

	resolved, cleanup, err := resolveBinaryScript(ferretsource.New(path, "RETURN 2"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resolved == path {
		t.Fatalf("expected a temporary snapshot, got original path %q", path)
	}

	content, err := os.ReadFile(resolved)
	if err != nil {
		t.Fatalf("failed to read snapshot: %v", err)
	}
	if string(content) != "RETURN 2" {
		t.Fatalf("unexpected snapshot content: %q", content)
	}

	if err := cleanup(); err != nil {
		t.Fatalf("expected no cleanup error, got %v", err)
	}
	if _, err := os.Stat(resolved); !os.IsNotExist(err) {
		t.Fatalf("expected snapshot to be removed, got %v", err)
	}
}
