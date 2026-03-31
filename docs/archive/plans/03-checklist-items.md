# Checklist Items

## Context

Todos can have checklist items. The CLI can set them on create/update (comma-separated via URL scheme), but can't read individual items or mark them complete. The GUI shows checklist items with checkboxes inside each todo.

## API Surface

**URL scheme (existing, on add/update):**
- `checklist-items` — set all items (newline-separated)
- `prepend-checklist-items`, `append-checklist-items` — on update

**Database (TMChecklistItem table):**
- Fields likely include: uuid, title, status, task (FK to TMTask), index
- Need to discover schema via `sqlite3 .schema TMChecklistItem`

**AppleScript:** The sdef does NOT expose checklist items. Things only supports them via URL scheme. No AppleScript read or write access.

**JXA:** No checklist item access documented.

## Commands to Build

```
things todo show <id>               # Already exists — enhance to show checklist items from DB
things todo show <id> --json        # Include checklist items in JSON output
```

Marking individual checklist items complete is likely impossible without the URL scheme `update` command replacing the entire checklist. This is a Things API limitation worth documenting.

## Implementation

### Step 1: Discover schema
```bash
sqlite3 <db-path> ".schema TMChecklistItem"
```

### Step 2: Add ChecklistItem type
```go
type ChecklistItem struct {
    ID     string `json:"id"`
    Name   string `json:"name"`
    Status Status `json:"status"`  // open or completed
}
```

### Step 3: Add DB query
```go
func (d *DB) ListChecklistItems(todoID string) ([]ChecklistItem, error)
```

### Step 4: Integrate into todo show
- `todo show <id>` displays checklist items below notes
- `todo show <id> --json` includes `checklistItems` array in output

### Step 5: Add to Todo type
```go
type Todo struct {
    // existing fields...
    ChecklistItems []ChecklistItem `json:"checklistItems,omitempty"`
}
```

## Open Questions

- Can individual checklist items be toggled via URL scheme without replacing the full list?
- Does `update?checklist-items=` replace or merge?

## Verification

```bash
# Create a todo with checklist items
./things todo add "Test" --checklist-items $'item 1\nitem 2\nitem 3'
# Show it
./things todo show <id>
./things todo show <id> --json
```
