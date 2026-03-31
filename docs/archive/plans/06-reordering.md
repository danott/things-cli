# Reordering (TUI)

## Context

Things allows drag-and-drop reordering in all views. The DB stores order via `"index"` (global position) and `todayIndex` (Today view position). The TUI will need vim-style motions to move items up/down.

## Primarily a TUI Feature

The CLI doesn't need reordering — it's a display tool. But the TUI will need:
- `j`/`k` to navigate
- `J`/`K` (or similar) to move items up/down in the list
- Persist new order back to Things

## API Surface for Writes

**URL scheme:** No reorder command.

**AppleScript:**
- `move to do "X" to beginning of list "Today"` — moves item within list
- `move to do "X" to before to do "Y"` — relative positioning
- Unclear if this works for arbitrary reordering within views

**Database:** Could potentially update `"index"` or `todayIndex` directly, but this violates the "never write to SQLite" rule.

## Open Questions

- Does AppleScript `move` command support fine-grained reordering within a view?
- Is there a way to set `todayIndex` via any documented API?
- Should we consider the JSON command for batch reorder operations?

## Research Needed

Test in Script Editor:
```applescript
tell application "Things3"
    set todayTodos to to dos of list "Today"
    -- Can we move item 5 to position 2?
    move item 5 of todayTodos to before item 2 of todayTodos
end tell
```

This plan is deferred until TUI development begins. The research above should be done first to confirm feasibility.
