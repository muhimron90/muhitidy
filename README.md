# muhitidy

`muhitidy` is a Go CLI utility for organizing files into predictable groups.

The project is deliberately conservative: it never assumes the current directory is the target, preview mode is the default, and existing files are never overwritten.

## Quick start

Install from the module :

```bash
go install github.com/muhimron90/muhitidy/cmd/muhitidy@latest
```

Preview an extension-based organization:

```bash
./muhitidy organize --target ~/Downloads
```

Actually apply it:

```bash
./muhitidy organize --target ~/Downloads --apply
```

Group by broad file categories:

```bash
./muhitidy organize --target ~/Downloads --by category --apply
```

Group by modification month:

```bash
./muhitidy organize --target ~/Downloads --by modified-month --apply
```

Recursively collect files from nested directories:

```bash
./muhitidy organize --target ~/Inbox --recursive --apply
```

## Grouping modes

| Mode             | Example destination    |
| ---------------- | ---------------------- |
| `extension`      | `pdf/report.pdf`       |
| `category`       | `Documents/report.pdf` |
| `modified-year`  | `2026/report.pdf`      |
| `modified-month` | `2026-10/report.pdf`   |
| `size`           | `1-10MB/report.pdf`    |

Extension grouping is case-insensitive. Files without an extension go to `no-extension`.

## Safety model

1. `--target` is mandatory.
2. No mutation happens unless `--apply` is explicitly supplied.
3. Only regular files are moved.
4. Symbolic links are skipped.
5. Hidden files and directories are skipped unless `--include-hidden` is supplied.
6. Existing destination files are never overwritten.
7. Conflicts default to deterministic numbered renames such as `report (1).pdf`.
8. A plan is computed before execution, so the preview describes the same set of moves that execution attempts.

## Conflict handling

`--on-conflict` supports:

- `rename` (default): choose the next available numbered filename.
- `skip`: leave the source untouched.
- `error`: fail planning instead of producing a partial plan.
