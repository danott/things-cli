# Logbook Pagination

## Context

The logbook query is hardcoded to `LIMIT 50` and `ORDER BY t.stopDate DESC`. For a user with years of completed todos, there's no way to browse history or filter by date range. JXA can't enumerate the full logbook (times out). The DB is the only viable path.

## Commands to Build

```
things todo list --logbook                    # existing, last 50
things todo list --logbook --limit 20         # control page size
things todo list --logbook --offset 50        # pagination
things todo list --logbook --since 2026-01-01 # date filter
things todo list --logbook --until 2026-03-01 # date filter
```

## Implementation

### Add flags to `todo_list.go`
- `--limit N` — default 50 for logbook
- `--offset N` — skip first N results
- `--since YYYY-MM-DD` — filter by completion date
- `--until YYYY-MM-DD` — filter by completion date

### Update DB query in `db.go`

Change `ListTodos` signature or add a `ListTodosOptions` struct:
```go
type ListOptions struct {
    Limit  int
    Offset int
    Since  *time.Time
    Until  *time.Time
}
```

The logbook query becomes:
```sql
AND t.status IN (2, 3)
AND t.stopDate >= ?    -- since (converted to Cocoa timestamp)
AND t.stopDate <= ?    -- until (converted to Cocoa timestamp)
GROUP BY t.uuid
ORDER BY t.stopDate DESC
LIMIT ? OFFSET ?
```

Note: `stopDate` is a Cocoa timestamp (seconds since 2001-01-01), not a Things date. Use `cocoaToTime`/inverse for conversion.

### Consider: Should --limit/--offset apply to all views?

Probably not by default — most views show all items. But having the option on logbook and trash (which can be large) makes sense. Could scope the flags to only those views initially.

## Verification

```bash
./things todo list --logbook --limit 10
./things todo list --logbook --limit 10 --offset 10
./things todo list --logbook --since 2026-03-01
./things todo list --logbook --since 2026-01-01 --until 2026-02-01
```
