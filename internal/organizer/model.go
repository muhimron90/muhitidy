package organizer

import "time"

type GroupBy string

const (
	GroupByExtension     GroupBy = "extension"
	GroupByCategory      GroupBy = "category"
	GroupByModifiedYear  GroupBy = "modified-year"
	GroupByModifiedMonth GroupBy = "modified-month"
	GroupBySize          GroupBy = "size"
)

func (g GroupBy) Valid() bool {
	switch g {
	case GroupByExtension, GroupByCategory, GroupByModifiedYear, GroupByModifiedMonth, GroupBySize:
		return true
	default:
		return false
	}
}

type ConflictMode string

const (
	ConflictRename ConflictMode = "rename"
	ConflictSkip   ConflictMode = "skip"
	ConflictError  ConflictMode = "error"
)

func (c ConflictMode) Valid() bool {
	switch c {
	case ConflictRename, ConflictSkip, ConflictError:
		return true
	default:
		return false
	}
}

type FileRecord struct {
	Path    string
	Name    string
	Size    int64
	ModTime time.Time
}

type Action string

const (
	ActionMove Action = "move"
	ActionSkip Action = "skip"
)

type Operation struct {
	Source      string `json:"source"`
	Destination string `json:"destination,omitempty"`
	Action      Action `json:"action"`
	Reason      string `json:"reason,omitempty"`
}

type Plan struct {
	Target       string      `json:"target"`
	GroupBy      GroupBy     `json:"group_by"`
	Operations   []Operation `json:"operations"`
	FilesScanned int         `json:"files_scanned"`
	MoveCount    int         `json:"move_count"`
	SkipCount    int         `json:"skip_count"`
}

func (p Plan) PlannedMoves() []Operation {
	moves := make([]Operation, 0, p.MoveCount)
	for _, op := range p.Operations {
		if op.Action == ActionMove {
			moves = append(moves, op)
		}
	}
	return moves
}

type ExecutionResult struct {
	Moved  int `json:"moved"`
	Failed int `json:"failed"`
}
