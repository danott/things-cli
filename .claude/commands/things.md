---
description: Manage Things 3 todos, projects, areas, and tags via things-cli
allowed-tools: Bash, Read, Write
---

You are helping the user manage their Things 3 task manager using the `things-cli` tool
(`go run .` from the repo root, or the `things` binary if built).

## Core output rules

- **Always use `--md`** for read commands. It is token-efficient and human-readable.
- Use `--json` **only** when you need structured data for a write operation (e.g., updating checklist items).
- Add `--verbose` / `-v` to any list command when you need IDs for subsequent writes.
- Never use `--gui` or `--interactive` unless the user explicitly asks to open Things.

## Read commands (always add --md)

```bash
go run . today --md              # today's list (default starting point)
go run . today --md --morning    # morning items only
go run . today --md --evening    # evening items only
go run . inbox --md              # unprocessed inbox
go run . upcoming --md           # scheduled future work
go run . anytime --md            # unscheduled available work
go run . someday --md            # deferred / maybe
go run . logbook --md --limit 20 # recent completed
go run . tomorrow --md

go run . todo show "title or ID" --md      # single todo detail
go run . project show "title or ID" --md  # single project detail
go run . todo list --project "Name" --md  # todos in a project
go run . todo list --area "Name" --md     # todos in an area
go run . todo list --tag "Name" --md      # todos by tag

go run . area list --md
go run . tag list --md
go run . project list --md
```

## Write commands (URL scheme — no auth token required)

Use these for all creates and most updates. They are fast and reliable.

```bash
# Create todo
go run . todo add "Title"
go run . todo add "Title" --when today --deadline 2026-04-15 --tags "Tag1,Tag2"
go run . todo add "Title" --list "Project Name" --heading "Section"
go run . todo add "Title" --notes "Freeform notes here"
go run . todo add "Title" --checklist-items "Step 1\nStep 2\nStep 3"

# Update todo (requires auth token)
go run . todo update <id> --title "New Title"
go run . todo update <id> --when tomorrow
go run . todo update <id> --append-notes "Additional context"
go run . todo update <id> --add-tags "NewTag"

# Lifecycle
go run . todo complete <id>
go run . todo cancel <id>
go run . todo delete <id>

# Projects
go run . project add "Title" --area "Area Name" --when today
go run . project complete <id>
go run . project delete <id>

# Areas & tags
go run . area add "Name"
go run . tag add "Name" --parent "Parent Tag"
```

## Updating checklist items (JSON API — preserves per-item completion state)

The URL scheme resets all checklist items to open. Use the JSON API when you need
to mark specific items complete without disturbing others.

**Workflow:**
1. Get current state: `go run . todo show <id> --json`
2. Write a JSON update file to `/tmp/things-update.json` with ALL checklist items
   (omitting any item removes it permanently):

```json
[{
  "type": "to-do",
  "id": "<todo-id>",
  "operation": "update",
  "attributes": {
    "checklist-items": [
      {"type": "checklist-item", "attributes": {"title": "Done item", "completed": true}},
      {"type": "checklist-item", "attributes": {"title": "Open item"}}
    ]
  }
}]
```

3. Apply: `go run . json --file /tmp/things-update.json`

Use `--dry-run` on any command to preview the URL scheme or AppleScript without executing.

## Workflow: process the inbox

```bash
go run . inbox --md -v    # -v surfaces IDs
# For each item: complete, reschedule, or move to a project
go run . todo update <id> --when today
go run . todo update <id> --list "Project Name"
go run . todo complete <id>
```

## Workflow: daily review

```bash
go run . today --md       # what's on deck
go run . upcoming --md    # what's coming
go run . anytime --md     # unscheduled pool
```

## Auth token

Some commands require an auth token (update, incomplete, duplicate):
- Store once: `go run . auth set`
- Or export: `export THINGS_AUTH_TOKEN=your-token`
- Check status: `go run . auth`

## Prefer --file over --data for special characters

When titles or notes contain backticks, quotes, or newlines, write JSON to a file and
use `--file` instead of `--data` to avoid shell quoting issues.
