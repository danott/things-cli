# Area CRUD

## Context

The CLI has zero area management. You can filter todos by area (`todo list --area`) and assign new projects/todos to areas, but you can't list, create, rename, or delete areas themselves. Areas have no URL scheme support — all writes must go through AppleScript.

## API Surface

**AppleScript (from Things.sdef):**
- `make new area with properties {name:"Work"}` — create
- `application.areas()` — list all
- `application.areas.byId("id")` — get by ID
- `set name of area "Old" to "New"` — rename
- `delete area "Work"` — delete
- Properties: `id` (r), `name` (rw), `collapsed` (rw), `tag names` (r)

**Database (TMArea table):**
- Already joined in `todosQuery` via `LEFT JOIN TMArea a ON t.area = a.uuid`
- Fields: `uuid`, `title`, plus others queryable via sqlite3

## Commands to Build

```
things area list                    # DB read (fast)
things area show <name-or-id>       # DB read — show area details + todo count
things area add <name>              # AppleScript: make new area
things area rename <old> <new>      # AppleScript: set name
things area delete <name>           # AppleScript: delete
```

## Implementation

### Domain layer (`internal/things/`)

**jxa.go** — add functions following the tag CRUD pattern:
```go
func CreateArea(name string) error
func RenameArea(oldName, newName string) error
func DeleteArea(name string) error
```

**db.go** — add:
```go
func (d *DB) ListAreas() ([]Area, error)    // SELECT uuid, title FROM TMArea ORDER BY "index"
func (d *DB) GetArea(id string) (*Area, error)
```

**types.go** — the `Area` struct already exists with ID and Name. No changes needed.

### Command layer (`internal/cmd/`)

**area.go** — new parent command with subcommands:
- `newAreaCmd()` — parent, adds list/show/add/rename/delete
- `newAreaListCmd()` — DB read
- `newAreaShowCmd()` — DB read, show name + todo count
- `newAreaAddCmd()` — AppleScript
- `newAreaRenameCmd()` — AppleScript
- `newAreaDeleteCmd()` — AppleScript

**root.go** — add `cmd.AddCommand(newAreaCmd())`

### Output (`internal/output/`)

**output.go** — add `PrintAreasText()` and `PrintAreasJSON()` if not already present.

## Verification

```bash
go vet ./...
./things area list
./things area add "TestArea" && ./things area list
./things area rename "TestArea" "Renamed" && ./things area list
./things area delete "Renamed" && ./things area list
```
