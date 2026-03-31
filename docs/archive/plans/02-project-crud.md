# Project CRUD (update, complete, cancel, delete)

## Context

Projects can be created (`project add`) and read (`project list`, `project show`), but can't be updated, completed, canceled, or deleted. The `update-project` URL scheme command exists but has no builder function. Project complete/cancel/delete need AppleScript (same pattern as todos).

## API Surface

**URL scheme (`things:///update-project`):**
- Required: `id`, `auth-token`
- Optional: `title`, `notes`, `prepend-notes`, `append-notes`, `when`, `deadline`, `tags`, `add-tags`, `area`, `area-id`, `completed`, `canceled`, `reveal`

**AppleScript (project inherits from to do):**
- `set status of project id "X" to completed`
- `set status of project id "X" to canceled`
- `delete project id "X"`
- `duplicate project id "X"`

## Commands to Build

```
things project update <id> [flags]  # URL scheme (requires auth)
things project complete <id>        # URL scheme: update-project?completed=true
things project cancel <id>          # URL scheme: update-project?canceled=true
things project delete <id>          # AppleScript (no URL scheme for delete)
```

## Implementation

### Domain layer (`internal/things/`)

**urlscheme.go** — add `UpdateProject` builder:
```go
func UpdateProject(id, authToken string, opts UpdateProjectOptions) *URLBuilder

type UpdateProjectOptions struct {
    Title       string
    Notes       string
    PrependNotes string
    AppendNotes  string
    When        string
    Deadline    string
    Tags        string
    AddTags     string
    Area        string
    AreaID      string
    Completed   bool
    Canceled    bool
    Reveal      bool
}
```

**jxa.go** — add delete function (complete/cancel go through URL scheme):
```go
func DeleteProject(id string) error  // AppleScript: delete project id "X"
```

**db.go** — enrich project queries to include area:
- `ListProjects()` should JOIN TMArea to include area name
- `GetProject()` same

**types.go** — add area fields to Project:
```go
type Project struct {
    ID       string
    Name     string
    Status   Status
    Notes    string
    AreaName string  // new
}
```

### Command layer (`internal/cmd/`)

**project_update.go** — new file, same pattern as `todo_update.go`:
- Flags for all UpdateProjectOptions fields
- Resolves auth token

**project_complete.go** — new file:
- `project complete <id>` — URL scheme with completed=true
- `project cancel <id>` — URL scheme with canceled=true
- `project delete <id>` — AppleScript

**project.go** — add subcommands

**root.go** — no changes needed (project cmd already registered)

## Verification

```bash
go vet ./...
./things project update <id> --title "New Name" --dry-run
./things project complete <id> --dry-run
./things project delete <id> --dry-run
# Live tests with real project IDs
```
