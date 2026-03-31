# Today: Morning/Evening Split

## Context

Things' Today view has a morning/evening split (configurable in Things settings). The DB already has this data — `startBucket` column: 0=morning, 1=evening. The Today query already sorts by `startBucket ASC` (morning first, then evening). But the CLI doesn't visually distinguish the two groups.

## Primarily a TUI Feature

For the CLI, the flat list sorted morning-then-evening is probably fine. But we should:

1. Expose `startBucket` in the data model so the TUI can render section headers
2. Optionally allow CLI filtering: `--morning` / `--evening`

## Implementation

### Add StartBucket to Todo scanning

**db.go** — `scanTodos` already scans `todayIndex` but not `startBucket`. Add it:
```go
// In scanTodos, add startBucket to the scan
// In Todo struct or a separate field
```

**types.go** — add field:
```go
type Todo struct {
    // existing fields...
    StartBucket int `json:"startBucket,omitempty"` // 0=morning, 1=evening (Today view only)
}
```

### CLI flags (optional)

**todo_list.go:**
```
things todo list --today              # all today items (current behavior)
things todo list --today --morning    # morning only
things todo list --today --evening    # evening only
```

### DB query variant

```sql
-- morning only
AND t.startBucket = 0

-- evening only
AND t.startBucket = 1
```

### Output

For text output, could add a separator line between morning and evening items. For JSON, the `startBucket` field is sufficient for consumers to group.

## Verification

```bash
./things todo list --today --json | jq '.[].startBucket'  # verify field present
./things todo list --today --morning
./things todo list --today --evening
```
