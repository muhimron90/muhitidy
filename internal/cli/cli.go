package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/muhimron90/muhitidy/internal/organizer"
)

const version = "0.1.0"

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}

	switch args[0] {
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	case "version", "--version":
		fmt.Fprintln(stdout, "muhitidy", version)
		return 0
	case "organize":
		if err := runOrganize(args[1:], stdout, stderr); err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 1
		}
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func printUsage(w io.Writer) {
	fmt.Fprint(w, `muhitidy organizes files into deterministic groups.

Usage:
  muhitidy organize --target PATH [options]
  muhitidy version

Safety:
  - --target is mandatory; there is no implicit current-directory target.
  - Preview mode is the default. Add --apply to actually move files.
  - Existing files are never overwritten.
  - Directories and symbolic links are not moved.

Organize options:
  --target PATH          Target directory to organize (required)
  --by MODE              extension, category, modified-year, modified-month, size
                         (default: extension)
  --apply                Execute the planned moves (default: preview only)
  --recursive            Include files in nested directories
  --include-hidden       Include dot-files and dot-directories
  --on-conflict MODE     rename, skip, error (default: rename)
  --timezone ZONE        local or UTC (default: local)
  --verbose              Show skipped files in the preview
  --json                 Print the plan/result as JSON
  --help                 Show command help

Examples:
  muhitidy organize --target ~/Downloads
  muhitidy organize --target ~/Downloads --by category
  muhitidy organize --target ~/Downloads --by modified-month --apply
  muhitidy organize --target ~/Inbox --recursive --on-conflict rename --apply
`)
}

func runOrganize(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("organize", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { printUsage(stderr) }

	target := fs.String("target", "", "target directory")
	by := fs.String("by", string(organizer.GroupByExtension), "grouping mode")
	apply := fs.Bool("apply", false, "apply planned moves")
	recursive := fs.Bool("recursive", false, "scan nested directories")
	includeHidden := fs.Bool("include-hidden", false, "include hidden files and directories")
	onConflict := fs.String("on-conflict", string(organizer.ConflictRename), "rename, skip, or error")
	timezone := fs.String("timezone", "local", "local or UTC")
	verbose := fs.Bool("verbose", false, "show skipped files")
	jsonOutput := fs.Bool("json", false, "print JSON output")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %s", strings.Join(fs.Args(), " "))
	}
	if strings.TrimSpace(*target) == "" {
		return errors.New("--target is required")
	}

	targetPath, err := filepath.Abs(*target)
	if err != nil {
		return fmt.Errorf("resolve target: %w", err)
	}
	info, err := os.Stat(targetPath)
	if err != nil {
		return fmt.Errorf("target %q: %w", targetPath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("target %q is not a directory", targetPath)
	}

	groupBy := organizer.GroupBy(*by)
	conflict := organizer.ConflictMode(*onConflict)
	if !groupBy.Valid() {
		return fmt.Errorf("invalid --by %q", *by)
	}
	if !conflict.Valid() {
		return fmt.Errorf("invalid --on-conflict %q", *onConflict)
	}
	loc, err := parseTimezone(*timezone)
	if err != nil {
		return err
	}

	files, err := organizer.Scan(targetPath, *recursive, *includeHidden)
	if err != nil {
		return err
	}
	plan, err := organizer.BuildPlan(files, organizer.PlannerOptions{
		Target:   targetPath,
		GroupBy:  groupBy,
		Conflict: conflict,
		Timezone: loc,
	})
	if err != nil {
		return err
	}

	if *apply {
		result, err := organizer.Execute(plan)
		if err != nil {
			if *jsonOutput {
				return writeJSON(stdout, map[string]any{
					"version": version,
					"mode":    "apply",
					"plan":    plan,
					"result":  result,
					"error":   err.Error(),
				})
			}
			return err
		}
		if *jsonOutput {
			return writeJSON(stdout, map[string]any{
				"version": version,
				"mode":    "apply",
				"plan":    plan,
				"result":  result,
			})
		}
		fmt.Fprintf(stdout, "Applied: moved %d file(s). Skipped %d.\n", result.Moved, plan.SkipCount)
		return nil
	}

	if *jsonOutput {
		return writeJSON(stdout, map[string]any{
			"version": version,
			"mode":    "preview",
			"plan":    plan,
		})
	}

	fmt.Fprintf(stdout, "Preview: %s\nGrouping: %s\nFiles scanned: %d\nPlanned moves: %d\nSkipped: %d\n\n", targetPath, groupBy, plan.FilesScanned, plan.MoveCount, plan.SkipCount)
	for _, op := range plan.Operations {
		if op.Action == organizer.ActionMove {
			fmt.Fprintf(stdout, "MOVE %q -> %q\n", op.Source, op.Destination)
		} else if *verbose {
			fmt.Fprintf(stdout, "SKIP %q (%s)\n", op.Source, op.Reason)
		}
	}
	if plan.MoveCount > 0 {
		fmt.Fprintln(stdout, "\nPreview only. Add --apply to execute these moves.")
	}
	return nil
}

func parseTimezone(value string) (*time.Location, error) {
	switch strings.ToLower(value) {
	case "local":
		return time.Local, nil
	case "utc":
		return time.UTC, nil
	default:
		return nil, fmt.Errorf("invalid --timezone %q; use local or UTC", value)
	}
}

func writeJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
