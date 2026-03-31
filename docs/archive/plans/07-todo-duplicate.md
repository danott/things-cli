# Todo Duplicate

## Context

Things URL scheme supports duplicating a todo via `update?duplicate=true`. This creates a copy and optionally applies modifications to it. Simple and consistent with the write API preference.

## Command

```
things todo duplicate <id>                    # duplicate as-is
things todo duplicate <id> --title "Copy of"  # duplicate and rename
```

## Implementation

### URL scheme

`things:///update?id=X&auth-token=T&duplicate=true` — duplicates the todo. Any other update params (title, notes, etc.) apply to the new copy, not the original.

### Domain layer

**urlscheme.go** — add `Duplicate` bool to `UpdateTodoOptions`:
```go
type UpdateTodoOptions struct {
    // existing fields...
    Completed bool
    Canceled  bool
    Duplicate bool  // new
}
```

Update `UpdateTodo` builder to set `duplicate=true` when `opts.Duplicate` is true.

### Command layer

**todo_duplicate.go** — new file:
```go
func newTodoDuplicateCmd() *cobra.Command
// Accepts optional --title, --notes, etc. to modify the copy
// Requires auth token (uses update URL scheme)
```

Or: add `--duplicate` flag to `todo update` and let the user compose:
```
things todo update <id> --duplicate --title "New copy"
```

The dedicated `todo duplicate <id>` subcommand is cleaner for the common case.

### Wire it up

**todo.go** — add `cmd.AddCommand(newTodoDuplicateCmd())`

## Verification

```bash
./things todo duplicate <id> --dry-run  # check URL
./things todo duplicate <id>            # verify copy appears in Things
```
