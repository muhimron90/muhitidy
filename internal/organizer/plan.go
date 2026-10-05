package organizer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

type PlannerOptions struct {
	Target   string
	GroupBy  GroupBy
	Conflict ConflictMode
	Timezone *time.Location
}

func BuildPlan(files []FileRecord, opts PlannerOptions) (Plan, error) {
	if !opts.GroupBy.Valid() {
		return Plan{}, fmt.Errorf("invalid grouping mode %q", opts.GroupBy)
	}
	if !opts.Conflict.Valid() {
		return Plan{}, fmt.Errorf("invalid conflict mode %q", opts.Conflict)
	}
	if opts.Timezone == nil {
		opts.Timezone = time.Local
	}

	occupied := make(map[string]struct{}, len(files)*2)
	for _, file := range files {
		occupied[pathKey(file.Path)] = struct{}{}
	}

	ops := make([]Operation, 0, len(files))
	for _, file := range files {
		group, err := Classify(file, opts.GroupBy, opts.Timezone)
		if err != nil {
			return Plan{}, err
		}
		destination := filepath.Join(opts.Target, group, file.Name)
		sourceKey := pathKey(file.Path)
		if sourceKey == pathKey(destination) {
			ops = append(ops, Operation{
				Source: file.Path,
				Action: ActionSkip,
				Reason: "already organized",
			})
			continue
		}

		resolved, conflict := resolveDestination(destination, occupied)
		if conflict {
			switch opts.Conflict {
			case ConflictSkip:
				ops = append(ops, Operation{Source: file.Path, Action: ActionSkip, Reason: "destination exists"})
				continue
			case ConflictError:
				return Plan{}, fmt.Errorf("destination conflict for %q: %q", file.Path, destination)
			case ConflictRename:
				destination = resolved
			}
		}

		occupied[pathKey(destination)] = struct{}{}
		ops = append(ops, Operation{
			Source:      file.Path,
			Destination: destination,
			Action:      ActionMove,
		})
	}

	sort.Slice(ops, func(i, j int) bool { return ops[i].Source < ops[j].Source })
	plan := Plan{
		Target:       opts.Target,
		GroupBy:      opts.GroupBy,
		Operations:   ops,
		FilesScanned: len(files),
	}
	for _, op := range ops {
		if op.Action == ActionMove {
			plan.MoveCount++
		} else {
			plan.SkipCount++
		}
	}
	return plan, nil
}

func resolveDestination(destination string, occupied map[string]struct{}) (string, bool) {
	_, occupiedByPlan := occupied[pathKey(destination)]
	_, existsOnDisk := os.Lstat(destination)
	conflict := occupiedByPlan || existsOnDisk == nil || !os.IsNotExist(existsOnDisk)
	if !conflict {
		return destination, false
	}

	ext := filepath.Ext(destination)
	stem := strings.TrimSuffix(filepath.Base(destination), ext)
	dir := filepath.Dir(destination)
	for i := 1; i < 1_000_000; i++ {
		candidate := filepath.Join(dir, stem+" ("+strconv.Itoa(i)+")"+ext)
		if _, ok := occupied[pathKey(candidate)]; ok {
			continue
		}
		if _, err := os.Lstat(candidate); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			continue
		}
		return candidate, true
	}
	return destination, true
}

func pathKey(path string) string {
	clean := filepath.Clean(path)
	if runtime.GOOS == "windows" {
		return strings.ToLower(clean)
	}
	return clean
}
