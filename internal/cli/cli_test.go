package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRequiresExplicitTarget(t *testing.T) {
	var out, errOut bytes.Buffer
	code := Run([]string{"organize"}, &out, &errOut)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(errOut.String(), "--target is required") {
		t.Fatalf("stderr = %q", errOut.String())
	}
}

func TestRunDefaultsToPreview(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "report.pdf"), []byte("pdf"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	code := Run([]string{"organize", "--target", dir}, &out, &errOut)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut.String())
	}
	if !strings.Contains(out.String(), "Preview only") {
		t.Fatalf("stdout = %q", out.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "report.pdf")); err != nil {
		t.Fatalf("source file missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "pdf", "report.pdf")); !os.IsNotExist(err) {
		t.Fatalf("destination unexpectedly exists: %v", err)
	}
}
