# things-cli

A command-line interface for [Cultured Code's Things 3](https://culturedcode.com/things/). Uses only documented APIs: the URL scheme for writes, AppleScript/JXA for some writes, and the SQLite database (read-only) for fast reads.

## Install

```bash
go install github.com/danott/things-cli@latest # not recommended in these early days
```

Or build from source:

```bash
git clone https://github.com/danott/things-cli
cd things-cli
go build -o things .
```

## Auth token

Some commands (`todo update`, `todo duplicate`, `project update`) require a Things auth token. Find yours in **Things > Settings > General > Enable Things URLs > Manage**.

Store it once in macOS Keychain:

```bash
things auth set
```

Or pass it per-command with `--auth-token`, or set the `THINGS_AUTH_TOKEN` environment variable. Priority: `--auth-token` flag > `THINGS_AUTH_TOKEN` env > Keychain.

Check current status:

```bash
things auth
```

## Global flags

These flags work on every command:

| Flag | Description |
|------|-------------|
| `--dry-run` | Print the action (URL or AppleScript) without executing it |
| `--json` | Output as JSON (on read commands) |
| `--md` | Output as Markdown (on read commands) |

## Views

Each view is a top-level subcommand that lists todos for that view.

```bash
things inbox
things today
things upcoming
things anytime
things someday
things logbook
things trash

# Open in Things.app instead of listing to stdout
things today --gui

# Show todo IDs alongside titles
things today --verbose

# Interactive mode — navigate and act on todos in the terminal
things today --interactive
things today -i

# Today morning/evening split
things today --morning
things today --evening

# Logbook pagination and filtering
things logbook --limit 20
things logbook --offset 40
things logbook --since 2026-01-01
things logbook --until 2026-03-01
things logbook --since 2026-01-01 --until 2026-03-31
```

## Todos

### List by project, area, or tag

```bash
things todo list --project "Website Redesign"
things todo list --area "Work"
things todo list --tag "Errands"

# Interactive mode
things todo list --project "Sprint 42" --interactive
things todo list --project "Sprint 42" -i
```

Exactly one of `--project`, `--area`, or `--tag` is required. Values can be a title or UUID.

### Show

```bash
things todo show <title_or_id>              # print details to stdout
things todo show <title_or_id> --json       # includes checklist items
things todo show <title_or_id> --gui        # open in Things.app
```

### Add

```bash
things todo add "Buy milk"
things todo add "Call doctor" --when today
things todo add "File taxes" --deadline 2026-04-15 --tags "Finance"
things todo add "Draft proposal" --notes "See Q2 brief" --list "Work"
things todo add "Groceries" --checklist-items $'Milk\nBread\nEggs'
things todo add "Review PR" --when today --list "Sprint 42" --heading "In Review"
things todo add "Buy milk" --dry-run        # print URL without opening
things todo add "Task" --reveal             # navigate to the created todo in Things
things todo add --show-quick-entry          # open Quick Entry dialog instead

# Open $EDITOR to compose the todo as markdown
things todo add --edit
things todo add "File taxes" --edit         # pre-populates the title
```

`--when` accepts: `today`, `tomorrow`, `evening`, `anytime`, `someday`, or `YYYY-MM-DD`.

#### Editor format

`--edit` opens `$VISUAL` or `$EDITOR` (fallback: `vi`) with a markdown document:

```markdown
---
when: today
deadline: 2026-04-15
tags: Finance, Work
list: Sprint 42
---

# File taxes

Notes as freeform prose.

- [ ] Gather W-2s
- [ ] Fill out forms
```

Frontmatter fields: `when`, `deadline`, `tags`, `list`, `list-id`, `heading`, `heading-id`. The body is free-form: the first `# H1` is the title, `- [ ]`/`- [x]` lines are checklist items, everything else is notes.

### Update

Requires auth token.

```bash
# Open $EDITOR with the current todo pre-populated
things todo update <id> --edit

things todo update <id> --title "New title"
things todo update <id> --when tomorrow
things todo update <id> --notes "Replace notes"
things todo update <id> --append-notes "Follow-up needed"
things todo update <id> --prepend-notes "Context: "
things todo update <id> --tags "Urgent,Work"           # replaces all tags
things todo update <id> --add-tags "Urgent"            # adds without removing existing
things todo update <id> --deadline 2026-06-01
things todo update <id> --list "Sprint 43"
things todo update <id> --list-id <project-id>
things todo update <id> --heading "In Review"
things todo update <id> --heading-id <heading-id>
things todo update <id> --checklist-items $'Step 1\nStep 2\nStep 3'   # replaces all items
things todo update <id> --append-checklist-items $'Step 4\nStep 5'
things todo update <id> --prepend-checklist-items "Step 0"
things todo update <id> --reveal
things todo update <id> --dry-run
```

### Complete, cancel, delete

```bash
things todo complete <id>
things todo cancel <id>
things todo delete <id>
```

### Duplicate

Requires auth token. Creates a copy; optional flags modify the copy.

```bash
things todo duplicate <id>
things todo duplicate <id> --title "Copy of original"
things todo duplicate <id> --when today --list "Inbox"
things todo duplicate <id> --dry-run
```

## Projects

### List and show

```bash
things project list
things project list --json
things project show <id>
things project show <id> --json
things project show <id> --gui
```

### Add

```bash
things project add "Website Redesign"
things project add "Sprint 42" --area "Work" --when today
things project add "Q2 Planning" --todos $'Research\nDraft\nReview'
things project add "Website Redesign" --dry-run
```

### Update

Requires auth token.

```bash
things project update <id> --title "New name"
things project update <id> --deadline 2026-06-30
things project update <id> --area "Personal"
things project update <id> --append-notes "Updated brief attached"
things project update <id> --dry-run
```

### Complete, cancel, delete

```bash
things project complete <id>
things project cancel <id>
things project delete <id>
```

## Areas

```bash
things area list
things area list --json
things area add "Health"
things area rename "Health" "Wellness"
things area delete "Wellness"
things area add "TestArea" --dry-run
```

## Tags

```bash
things tag list
things tag list --json
things tag add "Errand"
things tag add "Home" --parent "Places"
things tag rename "Errand" "Errands"
things tag delete "Errands"
```

## JSON command

Send a JSON array of create/update operations directly to Things. Full schema documented at [culturedcode.com](https://culturedcode.com/things/support/articles/2803573/).

```bash
# From file
things json --file plan.json

# Inline
things json --data '[{"type":"to-do","attributes":{"title":"Quick task","when":"today"}}]'

# From stdin
cat plan.json | things json

# Preview without executing
things json --file plan.json --dry-run

# Navigate to first created item
things json --file plan.json --reveal
```

Auth token is only required when the JSON contains `update` operations.

Example — create a project with headings and todos in one shot:

```json
[
  {
    "type": "project",
    "attributes": {
      "title": "Q2 Planning",
      "area": "Work",
      "items": [
        {"type": "heading", "attributes": {"title": "Research"}},
        {"type": "to-do", "attributes": {"title": "Market analysis", "when": "today"}},
        {"type": "to-do", "attributes": {"title": "Competitor review"}},
        {"type": "heading", "attributes": {"title": "Execution"}},
        {"type": "to-do", "attributes": {"title": "Draft proposal", "deadline": "2026-04-15"}}
      ]
    }
  }
]
```

## Custom actions

Register custom keybindings for interactive mode via `~/.config/things-cli/config.toml`. By default, actions receive the selected todo as JSON on stdin.

```toml
[[actions]]
key = "y"
label = "view"
command = "more"

[[actions]]
key = "p"
label = "copy id"
command = "pbcopy"
input = "id"
mode = "run"

[[actions]]
key = "s"
label = "process"
command = "my-script"
input = "id"
input_mode = "arg"
args = ["--verbose"]
# runs: my-script <todo-uuid> --verbose
```

| Field | Required | Description |
|-------|----------|-------------|
| `key` | yes | Single key to bind (must not conflict with built-in keys) |
| `label` | no | Shown in the help bar |
| `command` | yes | Executable to run |
| `args` | no | Static arguments passed to the command |
| `mode` | no | `exec` (default) takes over the terminal; `run` executes in the background and shows output in the status bar |
| `input` | no | What to pass: `json` (default), `markdown`, or `id` |
| `input_mode` | no | How to pass it: `stdin` (default) or `arg` (prepended to `args`) |

## Version

```bash
things version
```

## Architecture

- **Writes**: URL scheme (`things:///`) via `open -g` — fast, preferred for all writes. AppleScript fallback for operations the URL scheme doesn't support (area CRUD, tag CRUD, project/todo complete/cancel/delete).
- **Reads**: SQLite database (read-only, ~7ms) — primary read path. Never written to.
- **Auth token**: `--auth-token` flag > `THINGS_AUTH_TOKEN` env > macOS Keychain. Never stored in dotfiles.
