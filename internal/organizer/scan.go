package organizer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func Scan(target string, recursive, includeHidden bool) ([]FileRecord, error) {
	if !recursive {
		return scanTopLevel(target, includeHidden)
	}

	files := make([]FileRecord, 0)
	err := filepath.WalkDir(target, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == target {
			return nil
		}

		hidden := isHiddenName(d.Name())
		if hidden && !includeHidden {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("stat %q: %w", path, err)
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		files = append(files, FileRecord{
			Path:    path,
			Name:    d.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan %q: %w", target, err)
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

func scanTopLevel(target string, includeHidden bool) ([]FileRecord, error) {
	entries, err := os.ReadDir(target)
	if err != nil {
		return nil, fmt.Errorf("read directory %q: %w", target, err)
	}
	files := make([]FileRecord, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if isHiddenName(entry.Name()) && !includeHidden {
			continue
		}
		if entry.Type()&os.ModeSymlink != 0 {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("stat %q: %w", filepath.Join(target, entry.Name()), err)
		}
		if !info.Mode().IsRegular() {
			continue
		}
		files = append(files, FileRecord{
			Path:    filepath.Join(target, entry.Name()),
			Name:    entry.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

func isHiddenName(name string) bool {
	return strings.HasPrefix(name, ".") && name != "." && name != ".."
}
