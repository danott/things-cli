# things-cli: Research & Design Document

## 1. The API Surface: What Things Actually Exposes

Things 3 gives us three official interfaces. Understanding what each can and can’t do is the foundation of every design decision.

### URL Scheme (things:///)

The URL scheme is the primary write API. It supports six commands: `add`, `add-project`, `update`, `update-project`, `show`, `search`, plus the powerful `json` command and `version`.

Key capabilities:

- **add/add-project**: Create todos and projects with full metadata (title, notes, when, deadline, tags, checklist items, list assignment, headings)
- **update/update-project**: Modify existing items by ID. Requires `auth-token`. Supports `append-notes`, `prepend-notes`, `add-tags`, `append-checklist-items` — critical for non-destructive updates
- **show**: Navigate to any item, list, or built-in view by ID or query. Supports tag filtering via `filter` param
- **search**: Invoke the search UI
- **json**: The power command. Accepts an array of todo/project/heading/checklist-item objects. Supports both `create` and `update` operations in a single batch. Can create entire project hierarchies (project → headings → todos → checklist items) atomically

Constraints:

- `update` and `update-project` require the auth token
- The URL scheme is write-only for task data — you cannot *read* tasks through it
- Rate limit: 250 items per 10-second window
- Notes max 10,000 characters, strings max 4,000 characters
- The `json` command is the only way to create headings and structured projects in one shot
- x-callback-url support for getting created IDs back (via `x-success`)

### AppleScript

AppleScript is the read API (and a secondary write API). It provides programmatic access to Things’ object model.

Read capabilities:

- Access all built-in lists: Inbox, Today, Anytime, Upcoming, Someday, Logbook, Trash
- Iterate todos of any list, project, or area
- Get todo properties: name, notes, due date, creation date, modification date, completion date, cancellation date, status, tag names, id
- Get projects, areas, tags as collections
- Access selected todos in the UI
- The sort order matches Things’ UI — this is important for faithful TUI rendering

Write capabilities:

- Create todos, projects, areas, tags via `make` command
- Set/update all properties directly
- Move items between lists via `move` command
- Schedule items for specific dates via `schedule` command
- Complete/cancel items by setting `status`
- Delete items (moves to trash)
- Show quick entry panel (with autofill from other apps)
- Show/edit specific items in the UI

What AppleScript can do that URL scheme cannot:

- **Read** task data programmatically
- Delete items
- Create areas
- Create/manage tags
- Move items to specific lists
- Get selected items
- Access the Logbook for completed task history
- Access creation/modification dates

What URL scheme can do that AppleScript cannot:

- Create structured projects with headings in one shot (`json` command)
- Batch create/update via JSON
- Work cross-device (URL scheme works on iOS too, though we’re Mac-focused)

### SQLite Database (OFF LIMITS)

Several existing tools read the Things SQLite database directly. Per your constraint, we will not do this. The AppleScript API provides everything we need for reads, and it’s the *documented* interface — it won’t break with Things updates.

-----

## 2. Existing Tools: Landscape Assessment

I found ~8 existing Things CLI tools. Here’s the honest assessment:

### itspriddle/things-cli (Bash, 25 stars)

The cleanest existing implementation. Pure shell script that wraps the URL scheme. Commands mirror the URL scheme exactly: `things add`, `things update`, `things add-project`, `things update-project`, `things show`, `things search`. Includes man pages. This is the “reference implementation” that the Go port below explicitly cites.

**Strengths**: Clean, minimal, correct mapping of URL scheme. Has tests.
**Weaknesses**: Read-only via URL scheme (no listing tasks). Shell script limits extensibility. No TUI. No AppleScript integration for reads.

### ossianhempel/things3-cli (Go, 8 stars)

A Go rewrite explicitly modeled on itspriddle’s CLI. Adds database reads for listing tasks, AppleScript for deletes, and repeating task support (writes directly to DB for this). Uses `THINGS_AUTH_TOKEN` env var. Has `--dry-run` and `--foreground` flags.

**Strengths**: Go binary, good CLI patterns, env var for auth token, man pages. Adds listing via DB reads.
**Weaknesses**: Very new (6 commits), reads from the SQLite DB (violates your constraint), writes to DB for repeating tasks.

### thingsapi/things-cli (Python, popular)

Read-only CLI backed by the `things.py` library. Supports JSON/CSV/OPML/Gantt output. Good for data export and analysis. Commands like `things-cli today`, `things-cli inbox`, with filters for project/area/tag.

**Weaknesses**: Read-only, reads from SQLite, Python dependency.

### things3-cli (Rust, on crates.io)

A Rust implementation with an integrated MCP server. Full database access via async SQLx. Feature-flagged compilation.

**Weaknesses**: Reads from SQLite, heavy Rust toolchain, MCP focus over CLI UX.

### dan-hart/clings (Swift)

The most ambitious: natural language parsing for add (`clings add "buy milk tomorrow #errands"`), a TUI mode, stats, SQL-like search filters. Built with Swift/AppleScript.

**Strengths**: TUI with vim keys (j/k navigation, c to complete, Enter to open in Things), natural language parsing, shell completions.
**Weaknesses**: Swift-only (harder to contribute to), likely reads from DB for search.

### changkun/tli (Go)

Minimal. Inbox-only, sends via Things Cloud URL. Has rate limiting protection and content length checking.

### scriptingosx/ThingsCLITool (AppleScript-based)

Older, focused on basic operations via AppleScript directly.

### Assessment

None of these fully meet your standards:

- Most read from the SQLite database
- None combine URL scheme writes + AppleScript reads cleanly
- CLI grammar is inconsistent across the ecosystem
- No tool has both a great CLI *and* a great TUI
- MCP support exists separately but isn’t integrated into a CLI tool
- The `json` command’s power is largely untapped in existing CLIs

**The opportunity is a tool that uses only documented APIs (URL scheme + AppleScript), has consistent CLI grammar, includes a TUI, and is designed for both humans and AI agents.**

-----

## 3. Language & Framework Recommendation: Go

### Why Go

**Single binary distribution.** `brew install things` or `go install` and you’re done. No runtime dependencies. This matters enormously for a macOS-only tool — you don’t want to manage Ruby versions or Node.js for a task manager CLI.

**The Charm ecosystem.** Go has the best terminal UI libraries in the industry right now:

- **Bubble Tea**: Elm-architecture TUI framework. 18,000+ apps built with it. The `gh` CLI uses Charm’s libraries.
- **Lip Gloss**: CSS-like terminal styling
- **Bubbles**: Pre-built components (lists, text inputs, viewports, spinners, tables)
- **Huh**: Interactive form/prompt library (used by GitHub CLI for accessible prompts)
- **Gum**: Shell-scriptable interactive components

**Cobra for CLI parsing.** The industry standard for Go CLIs (`gh`, `kubectl`, `hugo` all use it). Gives you subcommands, flags, shell completions, and man page generation for free.

**AppleScript interop.** Go can shell out to `osascript` cleanly. The overhead is negligible for task management operations.

**Cross-compilation isn’t needed.** This is macOS-only by nature (Things is macOS/iOS only), so Go’s cross-platform story is irrelevant — but its build toolchain on macOS is excellent.

### Why not the alternatives

- **Ruby**: Your Jekyll familiarity is a plus, but gem management is a pain for end users, and the TUI ecosystem is weaker (TTY toolkit exists but isn’t Bubble Tea)
- **Swift**: Natural fit for macOS, but harder for others to contribute, and the CLI/TUI ecosystem is immature compared to Go
- **Node**: Runtime dependency, and the TUI libraries (ink) are less mature than Bubble Tea
- **Rust**: Great language, but the compile times and learning curve don’t justify it for a tool this focused. ratatui is good but Bubble Tea has more momentum
- **Bash**: itspriddle proved this works for writes, but it hits a wall for TUI and structured data handling

-----

## 4. CLI Design: Grammar & Consistency

Drawing from [clig.dev](https://clig.dev/), the `gh` CLI patterns, and your examples, here’s the proposed grammar.

### Core principle: noun-verb with consistent flags

```
things <noun> <verb> [positional] [--flags]
```

### Task operations

```bash
# Create
things todo add "Buy milk"
things todo add "Buy milk" --when today --tags "Errands" --list "Shopping"
things todo add "Buy milk" --notes "Organic, whole" --deadline 2026-04-01
things todo add "Call doctor" --when "next monday@9am"   # reminder
things todo add --titles "Milk\nBread\nCheese"           # bulk add

# Read
things todo list                        # defaults to Today
things todo list --inbox
things todo list --upcoming
things todo list --project "Vacation"
things todo list --area "Work" --tag "urgent"
things todo list --logbook --since "7 days ago"
things todo show <id>

# Update (requires auth token, stored in keychain or env)
things todo update <id> --title "Buy organic milk"
things todo update <id> --when tomorrow
things todo update <id> --append-notes "Check the farmers market"
things todo update <id> --add-tags "Health"
things todo complete <id>
things todo cancel <id>

# Delete
things todo delete <id>

# Open in Things
things todo show <id>          # prints to stdout (default)
things todo show <id> --gui    # opens in Things.app
things todo show <id> --tui    # opens in TUI, focused on this item
```

### Project operations

```bash
things project add "Website Redesign" --area "Work"
things project add "Sprint 42" --to-dos "Design\nImplement\nTest"
things project list
things project list --area "Work"
things project show <id>
things project show <id> --gui
things project show <id> --tui
things project update <id> --deadline 2026-06-01
things project complete <id>
things project delete <id>
```

### Tag operations

```bash
things tag list                                # all tags
things tag add "Errand"
things tag rename "Errand" "Errands"
things tag delete "Errands"
things tag parent "Home" --under "Places"      # set tag hierarchy
```

### Navigation / Views

```bash
things show today                  # prints today's tasks to stdout
things show today --gui            # navigates Things.app to Today
things show today --tui            # opens TUI on Today view
things show inbox
things show upcoming
things show anytime
things show someday
things show logbook
things show deadlines
```

### Batch operations via JSON

```bash
# The json command's full power, exposed cleanly
things json --file project.json
things json --data '[{"type":"to-do","attributes":{"title":"Test"}}]'
cat plan.json | things json --stdin
```

### Search

```bash
things search "quarterly report"
things search --tag "work" --project "Q1"
```

### Utility

```bash
things auth                # show token status, setup instructions
things auth --set          # store token in macOS Keychain
things version             # CLI version + Things version
things config              # show current config
```

### Reflection (your stretch goal)

```bash
things reflect today                           # show today's completed tasks with notes
things reflect week                            # this week's logbook
things reflect --project "Health" --since "30 days"
things reflect --tag "journal" --format markdown
```

### Design decisions embedded here

1. **Noun-verb not verb-noun.** `things todo add` not `things add`. This groups help text naturally and is consistent when you add project/area/tag nouns.
1. **Three output modes for `show` commands: stdout (default), `--gui`, `--tui`.** All show/list commands print to stdout by default — plain text that works for humans, pipes, and agents. `--gui` opens the item in Things.app. `--tui` opens the TUI focused on that item. Stdout is always the default because it’s the most composable. Inspired by `gh pr show --web`, but named honestly since Things has no web interface.
1. **Consistent flag names match Things’ own vocabulary.** `--when`, `--deadline`, `--tags`, `--list`, `--area`, `--notes`, `--append-notes` all come directly from the URL scheme parameter names. An agent reading the `--help` output should immediately map these to Things concepts.
1. **Positional argument for the common case.** `things todo add "Buy milk"` is the fast path. Everything else is a flag.
1. **Auth token in Keychain.** `security find-generic-password -s "things-cli" -a "auth-token" -w` is the macOS-native way. Fallback to `THINGS_AUTH_TOKEN` env var. Never store in a dotfile.

-----

## 5. The JSON Command: Deep Analysis

The `json` command is the most powerful part of the URL scheme and the key to making this CLI exceptional for agents.

### What it enables

A single `things json` call can:

- Create multiple projects, each with headings, todos, and checklist items
- Mix `create` and `update` operations in one batch
- Set all metadata (dates, tags, notes, completion status) per item
- Return created IDs via x-callback-url

### Why this matters for agents

An AI agent (Claude Code, an MCP server, a Shortcuts automation) can construct a complete project plan as JSON and send it in one shot:

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
        {"type": "to-do", "attributes": {"title": "Competitor review", "when": "tomorrow"}},
        {"type": "heading", "attributes": {"title": "Execution"}},
        {"type": "to-do", "attributes": {"title": "Draft proposal", "deadline": "2026-04-15"}}
      ]
    }
  }
]
```

### CLI design for JSON

```bash
# From file (most common for agents)
things json --file plan.json

# From stdin (for piping)
cat plan.json | things json

# Inline (for quick scripting)
things json --data '[{"type":"to-do","attributes":{"title":"Quick task"}}]'

# Dry run — print the URL without executing
things json --file plan.json --dry-run

# With reveal (navigate to first created item)
things json --file plan.json --reveal
```

The CLI should validate the JSON structure before sending it to Things, providing clear error messages about malformed objects. This is where a typed Go implementation shines over shell scripts.

-----

## 6. Auth Token Management

The auth token is required for `update`, `update-project`, and `json` (when updating). It must be:

1. **Easy to set up once**: `things auth --set` prompts and stores in macOS Keychain
1. **Predictable for agents**: `THINGS_AUTH_TOKEN` env var takes precedence
1. **Discoverable**: `things auth` shows status and setup instructions
1. **Secure**: Never written to a dotfile. Keychain + env var only.

Implementation:

```bash
# Store
security add-generic-password -s "things-cli" -a "auth-token" -w "$TOKEN"

# Retrieve  
security find-generic-password -s "things-cli" -a "auth-token" -w
```

Priority order: `--auth-token` flag > `THINGS_AUTH_TOKEN` env var > macOS Keychain.

-----

## 7. TUI Design

### Architecture

The TUI should be a separate entry point: `things tui` (or just `things` with no arguments, if we want it to be the default experience).

Built with Bubble Tea + Lip Gloss + Bubbles. The model:

```
┌─ Sidebar ──────┐┌─ Main List ──────────────────────────┐
│                ││                                       │
│ ▸ Inbox    (3) ││  Today                                │
│   Today    (5) ││  ─────                                │
│   Upcoming     ││  □ Review PR #123      Development    │
│   Anytime      ││  □ Buy groceries       Personal       │
│   Someday      ││  ■ Call dentist        Health         │
│                ││  □ Write blog post     Writing        │
│ ─ Areas ────── ││  □ Fix kitchen sink    Home           │
│ ▸ Work         ││                                       │
│ ▸ Personal     ││                                       │
│ ▸ Health       ││                                       │
│                ││                                       │
│ ─ Projects ─── ││                                       │
│   Sprint 42    ││                                       │
│   Vacation     ││                                       │
│                ││                                       │
└────────────────┘└───────────────────────────────────────┘
```

### Keyboard shortcuts — Things parity + Vim motions

The idea is dual-layer: Things’ own keyboard shortcuts work *and* Vim motions work. A Things user and a Vim user should both feel at home.

|Action            |Things shortcut|Vim motion|TUI binding|
|------------------|---------------|----------|-----------|
|Navigate down     |↓              |j         |Both work  |
|Navigate up       |↑              |k         |Both work  |
|Move to sidebar   |—              |h         |h or Tab   |
|Move to list      |—              |l         |l or Tab   |
|Expand task       |Enter          |Enter     |Enter      |
|Collapse / back   |Esc            |Esc       |Esc        |
|New todo          |⌘N             |o         |n or o     |
|Complete          |⌘K             |x         |x or c     |
|Today             |⌘1             |—         |1          |
|Upcoming          |⌘2             |—         |2          |
|Anytime           |⌘3             |—         |3          |
|Someday           |⌘4             |—         |4          |
|Search            |⌘F             |/         |/          |
|Open in Things.app|—              |—         |O          |
|Quick entry       |⌘N             |—         |n          |
|Quit              |⌘Q             |:q        |q          |
|Refresh           |⌘R             |—         |r          |

### Open in Things (the `--gui` escape hatch)

Pressing `O` on a task in the TUI opens it in Things.app via `things todo show <id> --gui`. Enter stays consistent with Things — it expands the task to show notes, checklist, tags, and dates. The TUI handles viewing; Things.app handles deep editing.

### Detail view

Enter expands a task inline, showing notes, checklist items, tags, and dates — just like Things. Esc collapses back to the list. For the reflection use case, this is where you’d read notes on completed tasks in the Logbook view.

-----

## 8. Agent & MCP Integration

### The CLI as an MCP-friendly tool

Rather than building a separate MCP server, the CLI itself should be agent-friendly:

1. **Structured output**: `--json` flag on all read commands. `things todo list --json` returns machine-parseable JSON. `--format` flag supports `json`, `csv`, `markdown`, `plain`.
1. **Predictable IDs**: All output includes Things IDs so agents can chain operations: list → filter → update.
1. **Dry run**: `--dry-run` on all write commands prints the URL or AppleScript that *would* execute, letting an agent preview before committing.
1. **Exit codes**: Standard Unix conventions. 0 = success, 1 = general error, 2 = usage error. Agents can check `$?`.
1. **Stdin support**: `things json --stdin` lets agents pipe generated JSON directly.

### Distributing skills

The CLI could ship with a `CLAUDE.md` or `AGENT.md` file that describes its capabilities for AI assistants. This file would include:

- Available commands with examples
- JSON schema for the `json` command
- Common workflows (e.g., “to add a task to today: `things todo add 'task' --when today`”)
- The flag vocabulary mapping to Things concepts

This is distributed alongside the binary. An agent with filesystem access reads the skill file; an MCP server references it in tool descriptions.

### MCP Server as a separate binary

For deeper integration, a `things-mcp` binary (built from the same Go codebase) could expose the CLI’s capabilities as MCP tools. Several community MCP servers already exist (hald/things-mcp, drjforrest/mcp-things3, hildersantos/things-mcp), but building it from the same codebase as the CLI ensures consistency and shared types.

-----

## 9. Reflection Feature (Stretch Goal)

This is the feature that makes this tool personally valuable beyond task management.

### The insight

Things already holds a rich history: every completed task with its notes, completion date, project context, and tags. The Logbook is an underused journal. If you leave notes on tasks as you work, you have a timestamped record of *what you did and what you thought about it*.

### Implementation

AppleScript gives us access to the Logbook. We can query:

- `to dos of list "Logbook"` — all completed tasks
- Filter by `completion date`, `project`, `area`, `tag names`
- Read `notes` for each task

```bash
# What did I do today?
things reflect today

# This week's work
things reflect week --area "Work"

# Monthly review of a project
things reflect --project "Health" --since "30 days"

# Export for journaling
things reflect week --format markdown > ~/journal/2026-w13.md

# Tag-based reflection (e.g., tasks tagged "journal" or "reflection")
things reflect --tag "reflection" --since "90 days"
```

### Output format for reflection

```markdown
## Friday, March 27, 2026

### Work
- ✓ Review Q2 budget proposal
  > Looks solid. Sarah's projections for APAC are conservative but defensible.
  > Need to follow up on headcount numbers.
- ✓ Ship v2.3 release notes

### Personal  
- ✓ Call Uncle Kenneth's family
  > Good conversation. They're doing okay.
```

The markdown output is designed to be readable as-is in the terminal, or piped into a file for a weekly review practice.

### Recurring task patterns

For recurring tasks (e.g., “Weekly review”), the reflection command could aggregate notes across instances, showing how your thinking evolves:

```bash
things reflect --title "Weekly review" --since "90 days" --recurring
```

This would show each instance’s notes chronologically — essentially a journal built from task completions.

-----

## 10. Project Structure

```
things-cli/
├── cmd/
│   └── things/
│       └── main.go
├── internal/
│   ├── cli/            # Cobra command definitions
│   │   ├── root.go
│   │   ├── todo.go     # things todo {add,list,show,update,complete,cancel,delete}
│   │   ├── project.go  # things project {add,list,show,update,complete,delete}
│   │   ├── show.go     # things show {today,inbox,...}
│   │   ├── search.go
│   │   ├── json_cmd.go # things json
│   │   ├── reflect.go  # things reflect
│   │   ├── auth.go     # things auth
│   │   └── tui.go      # things tui
│   ├── things/          # Core Things integration
│   │   ├── urlscheme.go # URL scheme builder & opener
│   │   ├── applescript.go # AppleScript execution & parsing
│   │   ├── auth.go      # Keychain + env var token management
│   │   ├── types.go     # Todo, Project, Area, Tag types
│   │   └── json.go      # JSON command builder & validator
│   ├── tui/             # Bubble Tea TUI
│   │   ├── app.go       # Main TUI model
│   │   ├── sidebar.go   # Sidebar component
│   │   ├── list.go      # Task list component
│   │   ├── preview.go   # Detail/preview pane
│   │   ├── keys.go      # Keybindings
│   │   └── styles.go    # Lip Gloss styles
│   └── output/          # Formatting
│       ├── table.go     # Terminal table output
│       ├── json.go      # JSON output
│       ├── markdown.go  # Markdown output (for reflect)
│       └── plain.go     # Plain text output
├── agent/
│   ├── CLAUDE.md        # Skill file for AI agents
│   └── schema.json      # JSON schema for the json command
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

-----

## 11. Distribution

```bash
# Homebrew (primary)
brew install things-cli

# Go install
go install github.com/yourusername/things-cli/cmd/things@latest

# From source
git clone ... && make install
```

Homebrew is the right primary channel for a macOS-only tool. The formula should install completions for zsh (macOS default), bash, and fish.

-----

## 12. Implementation Priority

### Phase 1: Core CLI (URL Scheme writes)

- `things todo add` with all flags
- `things project add`
- `things todo update` / `things project update`
- `things json` (file, stdin, data)
- `things auth` (keychain integration)
- `things show` (navigation commands)
- `things search`
- `things version`
- Shell completions (zsh, bash, fish)
- Man pages via Cobra

### Phase 2: AppleScript reads

- `things todo list` (all views: inbox, today, upcoming, etc.)
- `things project list`
- `things todo show <id>` (detail view)
- `--json` output on all read commands
- `things todo complete` / `cancel` / `delete` (via AppleScript)

### Phase 3: TUI

- Sidebar + list view
- Vim motions + Things shortcuts
- Enter to open in Things.app
- Quick complete/cancel
- Search

### Phase 4: Reflection

- `things reflect` with date ranges
- Project/area/tag filtering
- Markdown export
- Recurring task aggregation

### Phase 5: Agent integration

- `CLAUDE.md` skill file
- JSON schema for `json` command
- `things-mcp` binary (optional, same codebase)

-----

## 13. Resolved Design Decisions

1. **`things` with no args shows help.** The TUI is an explicit opt-in via `things tui`. Unlike lazygit (which *is* a TUI), this tool has both CLI commands and a TUI. Cobra convention applies: no args = help. Agents and scripts get useful output instead of an interactive session that blocks.
1. **Noun-verb only in v1, no shortcuts.** `things todo add "Buy milk"`, not `things add "Buy milk"`. Consistency wins. Shortcuts like `things add` can be added later, or users can establish their own shell aliases.
1. **ID retrieval is a goal, implementation TBD.** Getting created IDs back is important for scripting and agent workflows. The best mechanism (x-callback-url temporary server, AppleScript query, or parsing output) will be determined during Phase 1 implementation when we can feel the actual friction.
1. **TUI uses filesystem watching on the SQLite file, reads via AppleScript.** Watch the Things SQLite file with `fsnotify` (macOS kqueue/FSEvents) for modification timestamps, then trigger an AppleScript read when it changes. The database file is the *signal*, AppleScript is the *API*. We never parse a byte of SQLite. This gives near-real-time updates without polling overhead, and the boundary is clean: we observe that the file changed, we never reach into it.
1. **Binary is named `things`.** Until proven conflicting. Simple, memorable, what you’d expect to type.
