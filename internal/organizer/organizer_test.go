package organizer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestClassify(t *testing.T) {
	t.Parallel()
	file := FileRecord{Name: "Report.PDF", Size: 2 * 1024 * 1024, ModTime: time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)}
	if got := mustClassify(t, file, GroupByExtension); got != "pdf" {
		t.Fatalf("extension = %q", got)
	}
	if got := mustClassify(t, file, GroupByCategory); got != "Documents" {
		t.Fatalf("category = %q", got)
	}
	if got := mustClassify(t, file, GroupByModifiedYear); got != "2026" {
		t.Fatalf("year = %q", got)
	}
	if got := mustClassify(t, file, GroupByModifiedMonth); got != "2026-10" {
		t.Fatalf("month = %q", got)
	}
	if got := mustClassify(t, file, GroupBySize); got != "1-10MB" {
		t.Fatalf("size = %q", got)
	}
}

func TestExtensionGroup(t *testing.T) {
	cases := map[string]string{
		"photo.JPG":      "jpg",
		"archive.tar.gz": "gz",
		"README":         "no-extension",
		".env":           "no-extension",
	}
	for name, want := range cases {
		if got := extensionGroup(name); got != want {
			t.Errorf("extensionGroup(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestBuildPlanRenamesConflicts(t *testing.T) {
	dir := t.TempDir()
	existingDir := filepath.Join(dir, "pdf")
	if err := os.MkdirAll(existingDir, 0o755); err != nil {
		t.Fatal(err)
	}
	existing := filepath.Join(existingDir, "report.pdf")
	if err := os.WriteFile(existing, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	files := []FileRecord{{Path: filepath.Join(dir, "report.pdf"), Name: "report.pdf"}}
	plan, err := BuildPlan(files, PlannerOptions{Target: dir, GroupBy: GroupByExtension, Conflict: ConflictRename})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Operations) != 1 {
		t.Fatalf("operations = %d", len(plan.Operations))
	}
	want := filepath.Join(existingDir, "report (1).pdf")
	if got := plan.Operations[0].Destination; got != want {
		t.Fatalf("destination = %q, want %q", got, want)
	}
}

func TestExecuteMovesFiles(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "photo.jpg")
	if err := os.WriteFile(source, []byte("image"), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := Scan(dir, false, false)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := BuildPlan(files, PlannerOptions{Target: dir, GroupBy: GroupByExtension, Conflict: ConflictRename})
	if err != nil {
		t.Fatal(err)
	}
	result, err := Execute(plan)
	if err != nil {
		t.Fatal(err)
	}
	if result.Moved != 1 {
		t.Fatalf("moved = %d", result.Moved)
	}
	moved := filepath.Join(dir, "jpg", "photo.jpg")
	if _, err := os.Stat(moved); err != nil {
		t.Fatalf("moved file missing: %v", err)
	}
}

func TestScanSkipsHiddenAndSymlinks(t *testing.T) {
	dir := t.TempDir()
	mustWrite := func(name string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite("visible.txt")
	mustWrite(".hidden.txt")
	if err := os.Symlink(filepath.Join(dir, "visible.txt"), filepath.Join(dir, "link.txt")); err != nil && !strings.Contains(err.Error(), "not supported") {
		t.Fatal(err)
	}

	files, err := Scan(dir, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Name != "visible.txt" {
		t.Fatalf("unexpected files: %+v", files)
	}
}

func mustClassify(t *testing.T, file FileRecord, by GroupBy) string {
	t.Helper()
	got, err := Classify(file, by, time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	return got
}
