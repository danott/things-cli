# Feature Plans

Each plan is self-contained with enough context for independent implementation. Priority order:

1. [Area CRUD](01-area-crud.md) — list, add, rename, delete areas (AppleScript + DB)
2. [Project CRUD](02-project-crud.md) — update, complete, cancel, delete projects (URL scheme + AppleScript)
3. [Checklist Items](03-checklist-items.md) — read/display checklist items from DB
4. [Logbook Pagination](04-logbook-pagination.md) — limit, offset, date filtering
5. [Today Morning/Evening](05-today-morning-evening.md) — expose startBucket for TUI
6. [Reordering](06-reordering.md) — TUI vim motions (research needed, deferred)
7. [Todo Duplicate](07-todo-duplicate.md) — duplicate via URL scheme

## Conventions

- **Writes**: URL scheme first, AppleScript fallback (see CLAUDE.md)
- **Reads**: SQLite DB (fast), validated against JXA (ground truth) via integration tests
- **Commands**: noun-verb grammar (`things <noun> <verb>`)
- **Flags**: `--dry-run` on all write commands, `--json` on all read commands
- **Tests**: `go test -tags=integration ./internal/things/` for DB vs JXA validation
