# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build -o things .               # build binary
go run . <subcommand>               # run without building
go test ./...                       # run unit tests
go test -tags=integration -v ./internal/things/  # integration tests (requires Things 3 running)
```

The binary requires macOS — it reads the Things 3 SQLite database and uses `osascript` (JXA) and `open` (URL scheme).

CGO is required (`CGO_ENABLED=1`) because of the `mattn/go-sqlite3` dependency.

## Architecture

Three write/read paths, chosen deliberately:

- **URL scheme** (`things:///` via `open -g`): preferred write path. Fast, officially supported by Things. Built in `internal/things/urlscheme.go`.
- **JXA/AppleScript** (`osascript -l JavaScript`): fallback for operations the URL scheme doesn't support (area CRUD, tag CRUD, complete/cancel/delete). Built in `internal/things/jxa.go` with script templates in `jxa_scripts.go`.
- **SQLite** (read-only, `~7ms`): all reads. Never written to. The database lives in `~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac`. Built in `internal/things/db.go`.

**Auth token** resolution order: `--auth-token` flag > `THINGS_AUTH_TOKEN` env > macOS Keychain. Managed in `internal/things/auth.go`.

## Package layout

- `internal/cmd/` — Cobra command tree. One file per subcommand. `root.go` wires everything together.
- `internal/things/` — Core domain: DB reads, URL scheme builder, JXA runner, auth, types, markdown parsing.
- `internal/output/` — Shared formatting (plain text vs JSON).
- `internal/tui/` — Bubbletea TUI (sidebar, todo list, detail view).

## Using the CLI to manage Things

The CLI can manage its own task tracking in Things. Common operations:

```bash
# Find a todo
go run . search "things-cli"
go run . today show

# Show a todo with checklist items
go run . todo show <id> --json

# Complete a todo
go run . todo complete <id>

# Complete checklist items (requires replacing ALL items with desired states)
# Write JSON to a file, then pass with --file (avoids shell quoting issues):
go run . json --file /path/to/update.json
```

### Completing checklist items

Checklist items cannot be updated individually. The Things JSON API requires sending **all checklist items** as a nested attribute of a todo update. Any items not included will be removed.

```json
[{
  "type": "to-do",
  "id": "<todo-id>",
  "operation": "update",
  "attributes": {
    "checklist-items": [
      {"type": "checklist-item", "attributes": {"title": "First item", "completed": true}},
      {"type": "checklist-item", "attributes": {"title": "Second item"}}
    ]
  }
}]
```

Key details:
- `"completed": true` marks an item done; omit the field (or `false`) for open items
- You must include **every** checklist item — this is a full replacement, not a patch
- Use `--file` instead of `--data` when titles contain backticks or special characters
- Use `--dry-run` to preview the URL without executing
- The `things json` command auto-injects the auth token

## Testing

Integration tests (`-tags=integration`) compare DB query results against JXA output to verify SQLite queries match what the Things GUI shows. They require Things 3 to be running.
