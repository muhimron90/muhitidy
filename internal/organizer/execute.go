package organizer

import (
	"fmt"
	"os"
	"path/filepath"
)

func Execute(plan Plan) (ExecutionResult, error) {
	result := ExecutionResult{}
	for _, op := range plan.Operations {
		if op.Action != ActionMove {
			continue
		}
		if _, err := os.Lstat(op.Source); err != nil {
			result.Failed++
			return result, fmt.Errorf("source changed before move %q: %w", op.Source, err)
		}
		if _, err := os.Lstat(op.Destination); err == nil {
			result.Failed++
			return result, fmt.Errorf("destination appeared before move %q", op.Destination)
		} else if !os.IsNotExist(err) {
			result.Failed++
			return result, fmt.Errorf("check destination %q: %w", op.Destination, err)
		}

		if err := os.MkdirAll(filepath.Dir(op.Destination), 0o755); err != nil {
			result.Failed++
			return result, fmt.Errorf("create destination directory %q: %w", filepath.Dir(op.Destination), err)
		}
		if err := os.Rename(op.Source, op.Destination); err != nil {
			result.Failed++
			return result, fmt.Errorf("move %q to %q: %w", op.Source, op.Destination, err)
		}
		result.Moved++
	}
	return result, nil
}
