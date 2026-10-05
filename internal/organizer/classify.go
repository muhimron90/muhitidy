package organizer

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

func Classify(file FileRecord, by GroupBy, loc *time.Location) (string, error) {
	switch by {
	case GroupByExtension:
		return extensionGroup(file.Name), nil
	case GroupByCategory:
		return categoryGroup(file.Name), nil
	case GroupByModifiedYear:
		return file.ModTime.In(loc).Format("2006"), nil
	case GroupByModifiedMonth:
		return file.ModTime.In(loc).Format("2006-01"), nil
	case GroupBySize:
		return sizeGroup(file.Size), nil
	default:
		return "", fmt.Errorf("unsupported grouping mode %q", by)
	}
}

func extensionGroup(name string) string {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(name)), ".")
	if ext == "" || strings.HasPrefix(name, ".") && !strings.Contains(strings.TrimPrefix(name, "."), ".") {
		return "no-extension"
	}
	return ext
}

func categoryGroup(name string) string {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(name)), ".")
	if ext == "" {
		return "Other"
	}
	for category, extensions := range categoryExtensions {
		if extensions[ext] {
			return category
		}
	}
	return "Other"
}

var categoryExtensions = map[string]map[string]bool{
	"Images": {
		"jpg": true, "jpeg": true, "png": true, "gif": true, "webp": true,
		"bmp": true, "tif": true, "tiff": true, "svg": true, "ico": true,
	},
	"Videos": {
		"mp4": true, "m4v": true, "mov": true, "mkv": true, "avi": true,
		"webm": true, "wmv": true, "flv": true, "mpeg": true, "mpg": true,
	},
	"Audio": {
		"mp3": true, "wav": true, "flac": true, "aac": true, "m4a": true,
		"ogg": true, "opus": true, "wma": true,
	},
	"Documents": {
		"pdf": true, "doc": true, "docx": true, "odt": true, "rtf": true,
		"txt": true, "md": true, "csv": true, "tsv": true, "xls": true,
		"xlsx": true, "ods": true, "ppt": true, "pptx": true, "odp": true,
	},
	"Archives": {
		"zip": true, "7z": true, "rar": true, "tar": true, "gz": true,
		"bz2": true, "xz": true, "zst": true, "tgz": true, "iso": true,
	},
	"Code": {
		"go": true, "rs": true, "c": true, "h": true, "cc": true, "cpp": true,
		"cxx": true, "hpp": true, "java": true, "kt": true, "kts": true,
		"py": true, "js": true, "jsx": true, "ts": true, "tsx": true,
		"html": true, "css": true, "scss": true, "sql": true, "sh": true,
		"ps1": true, "bat": true, "cmd": true, "toml": true, "yaml": true,
		"yml": true, "json": true, "xml": true,
	},
}

func sizeGroup(size int64) string {
	const mib = int64(1024 * 1024)
	switch {
	case size == 0:
		return "Empty"
	case size < 1*mib:
		return "Under-1MB"
	case size < 10*mib:
		return "1-10MB"
	case size < 100*mib:
		return "10-100MB"
	case size < 1024*mib:
		return "100MB-1GB"
	default:
		return "Over-1GB"
	}
}
