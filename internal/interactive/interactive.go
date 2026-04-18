package interactive

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"

	"github.com/danott/things-cli/internal/config"
	"github.com/danott/things-cli/internal/editor"
	"github.com/danott/things-cli/internal/things"
)

var (
	cursorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	doneStyle      = lipgloss.NewStyle().Strikethrough(true).Foreground(lipgloss.Color("8"))
	cancelStyle    = lipgloss.NewStyle().Strikethrough(true).Foreground(lipgloss.Color("1"))
	dimStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	statusBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	boldStyle      = lipgloss.NewStyle().Bold(true)
)

type model struct {
	view      string
	title     string
	loader    func() ([]things.Todo, error)
	db        *things.DB
	items     []things.Todo
	cursor    int
	offset    int // scroll offset for viewport
	height    int // terminal height
	width     int // terminal width
	adding    bool
	addInput  textinput.Model
	authToken string
	err       error
	status    string
	actions   []config.Action
	addList   string // list context for adding todos (project/area name); empty = use view default

	// Detail view state
	detailOpen     bool
	detailMarkdown string   // raw markdown; source of truth for re-wrapping on resize
	detailLines    []string // word-wrapped lines derived from detailMarkdown
	detailOffset   int
	singleMode     bool // when true, quitting detail view quits the app
}

type dbChangedMsg struct{}
type statusClearMsg struct{}
type editorFinishedMsg struct {
	todoID   string
	tempPath string
	err      error
}

type actionFinishedMsg struct {
	err error
}

func Run(db *things.DB, view, authToken string, actions []config.Action) error {
	loader := func() ([]things.Todo, error) {
		return db.ListTodosWithCompleted(view)
	}
	return RunWithLoader(db, view, view, loader, authToken, actions, "")
}

// RunWithLoader starts the interactive TUI with a custom todo loader.
// title is shown as the list heading; view is used for add/when behavior.
// addList, when non-empty, enables adding todos and sets the list they go into.
func RunWithLoader(db *things.DB, title, view string, loader func() ([]things.Todo, error), authToken string, actions []config.Action, addList string) error {
	m, err := newModel(db, title, view, loader, authToken, actions)
	if err != nil {
		return err
	}
	m.addList = addList
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

// RunSingle starts the interactive TUI showing the detail view for a single todo.
func RunSingle(db *things.DB, todoID, authToken string, actions []config.Action) error {
	todo, err := db.GetTodo(todoID)
	if err != nil {
		return err
	}

	loader := func() ([]things.Todo, error) {
		t, err := db.GetTodo(todoID)
		if err != nil {
			return nil, err
		}
		return []things.Todo{*t}, nil
	}

	ti := textinput.New()
	ti.Prompt = "  + "
	ti.Placeholder = "new todo"
	ti.CharLimit = 256

	m := model{
		view:         "",
		title:        todo.Name,
		loader:       loader,
		db:           db,
		items:        []things.Todo{*todo},
		addInput:     ti,
		authToken:    authToken,
		actions:      actions,
		height:       24,
		detailOpen:     true,
		detailMarkdown: things.TodoToMarkdown(todo),
		detailLines:    wrapDetailContent(things.TodoToMarkdown(todo), 0),
		detailOffset:   0,
		singleMode:   true,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func newModel(db *things.DB, title, view string, loader func() ([]things.Todo, error), authToken string, actions []config.Action) (model, error) {
	todos, err := loader()
	if err != nil {
		return model{}, err
	}

	ti := textinput.New()
	ti.Prompt = "  + "
	ti.Placeholder = "new todo"
	ti.CharLimit = 256

	return model{
		view:      view,
		title:     title,
		loader:    loader,
		db:        db,
		items:     todos,
		addInput:  ti,
		authToken: authToken,
		actions:   actions,
		height:    24, // default, updated on WindowSizeMsg
	}, nil
}

func (m model) Init() tea.Cmd {
	return watchDB()
}

// listHelpSegments returns the ordered help segments for the list view.
func (m model) listHelpSegments() []string {
	segs := []string{"j/k: navigate", "enter: view", "e: edit", "o: reveal", "x: complete", "ctrl+x: cancel", "X: clear done"}
	if m.canAdd() {
		segs = append(segs, "a: add")
	}
	for _, action := range m.actions {
		segs = append(segs, action.Key+": "+action.Label)
	}
	return append(segs, "q: quit")
}

// listHeight returns how many item rows fit in the viewport.
func (m model) listHeight() int {
	helpLines := strings.Count(wrapHelp(m.listHelpSegments(), m.width), "\n") + 1
	reserved := 3 + helpLines // title + blank + blank-before-help + helpLines
	if m.adding {
		reserved++
	}
	if m.status != "" {
		reserved++
	}
	if m.err != nil {
		reserved++
	}
	h := m.height - reserved
	if h < 1 {
		h = 1
	}
	return h
}

func (m *model) clampScroll() {
	lh := m.listHeight()
	// Ensure cursor is visible
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+lh {
		m.offset = m.cursor - lh + 1
	}
	// Clamp offset
	if m.offset < 0 {
		m.offset = 0
	}
	max := len(m.items) - lh
	if max < 0 {
		max = 0
	}
	if m.offset > max {
		m.offset = max
	}
}

// detailHelpSegments returns the ordered help segments for the detail view.
func (m model) detailHelpSegments() []string {
	segs := []string{"j/k: scroll", "x: complete", "ctrl+x: cancel", "c: copy", "e: edit", "o: reveal"}
	for _, action := range m.actions {
		segs = append(segs, action.Key+": "+action.Label)
	}
	if m.singleMode {
		return append(segs, "q: quit")
	}
	return append(segs, "enter/esc: back")
}

// detailHeight returns how many content lines fit in the detail viewport.
func (m model) detailHeight() int {
	helpLines := strings.Count(wrapHelp(m.detailHelpSegments(), m.width), "\n") + 1
	h := m.height - 1 - helpLines // blank-before-help + helpLines
	if h < 1 {
		h = 1
	}
	return h
}

func (m *model) clampDetailScroll() {
	max := len(m.detailLines) - m.detailHeight()
	if max < 0 {
		max = 0
	}
	if m.detailOffset > max {
		m.detailOffset = max
	}
	if m.detailOffset < 0 {
		m.detailOffset = 0
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		if m.detailOpen {
			if m.detailMarkdown != "" {
				m.detailLines = wrapDetailContent(m.detailMarkdown, m.width)
			}
			m.clampDetailScroll()
		} else {
			m.clampScroll()
		}
		return m, nil

	case dbChangedMsg:
		m = m.refresh()
		return m, watchDB()

	case statusClearMsg:
		m.status = ""
		return m, nil

	case editorFinishedMsg:
		defer os.Remove(msg.tempPath)
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		if err := m.applyEdit(msg.todoID, msg.tempPath); err != nil {
			m.err = err
		}
		return m, nil

	case actionFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		return m, nil

	case tea.KeyMsg:
		if m.adding {
			return m.updateAdding(msg)
		}
		if m.detailOpen {
			return m.updateDetail(msg)
		}
		return m.updateNormal(msg)
	}

	return m, nil
}

func (m model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		return m, tea.Quit

	case "j", "down":
		if m.cursor < len(m.items)-1 {
			m.cursor++
			m.clampScroll()
		}

	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
			m.clampScroll()
		}

	case "g", "home":
		m.cursor = 0
		m.clampScroll()

	case "G", "end":
		if len(m.items) > 0 {
			m.cursor = len(m.items) - 1
			m.clampScroll()
		}

	case "x":
		if m.cursor < len(m.items) {
			todo := m.items[m.cursor]
			switch todo.Status {
			case things.StatusCompleted:
				b := things.UpdateTodo(todo.ID, m.authToken, things.UpdateTodoOptions{Status: things.StatusPtr(things.StatusOpen)})
				_ = b.Open()
			case things.StatusOpen:
				b := things.UpdateTodo(todo.ID, m.authToken, things.UpdateTodoOptions{Status: things.StatusPtr(things.StatusCompleted)})
				if err := b.Open(); err != nil {
					m.err = err
				}
			}
		}

	case "ctrl+x":
		if m.cursor < len(m.items) {
			todo := m.items[m.cursor]
			switch todo.Status {
			case things.StatusCanceled:
				b := things.UpdateTodo(todo.ID, m.authToken, things.UpdateTodoOptions{Status: things.StatusPtr(things.StatusOpen)})
				_ = b.Open()
			case things.StatusOpen:
				b := things.UpdateTodo(todo.ID, m.authToken, things.UpdateTodoOptions{Status: things.StatusPtr(things.StatusCanceled)})
				if err := b.Open(); err != nil {
					m.err = err
				}
			}
		}

	case "X":
		if err := things.LogCompleted(); err != nil {
			m.err = err
		}
		// DB watcher will pick up the change and refresh

	case "enter":
		if m.cursor < len(m.items) {
			todo, err := m.db.GetTodo(m.items[m.cursor].ID)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.setDetailContent(things.TodoToMarkdown(todo))
			m.detailOffset = 0
			m.detailOpen = true
		}

	case "e":
		if m.cursor < len(m.items) {
			return m.openEditor()
		}

	case "o":
		if m.cursor < len(m.items) {
			todo := m.items[m.cursor]
			if err := things.ShowItem(todo.ID).OpenForeground(); err != nil {
				m.err = err
			}
		}

	case "a":
		if m.canAdd() {
			m.adding = true
			m.addInput.Reset()
			m.addInput.Focus()
			return m, textinput.Blink
		}

	default:
		if m.cursor < len(m.items) {
			for _, action := range m.actions {
				if msg.String() == action.Key {
					return m.runAction(action)
				}
			}
		}
	}

	return m, nil
}

func (m model) updateAdding(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.adding = false
		return m, nil

	case "enter":
		title := strings.TrimSpace(m.addInput.Value())
		if title == "" {
			m.adding = false
			return m, nil
		}
		m.adding = false
		opts := things.AddTodoOptions{When: viewToWhen(m.view), List: m.addList}
		b := things.AddTodo(title, opts)
		if err := b.Open(); err != nil {
			m.err = err
		} else {
			m.status = fmt.Sprintf("Added: %s", title)
		}
		return m, clearStatusAfter(2 * time.Second)

	default:
		var cmd tea.Cmd
		m.addInput, cmd = m.addInput.Update(msg)
		return m, cmd
	}
}

func (m model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		if m.singleMode {
			return m, tea.Quit
		}
		m.detailOpen = false
		return m, nil

	case "enter":
		if m.singleMode {
			return m, tea.Quit
		}
		m.detailOpen = false
		return m, nil

	case "j", "down":
		m.detailOffset++
		m.clampDetailScroll()

	case "k", "up":
		m.detailOffset--
		m.clampDetailScroll()

	case "g", "home":
		m.detailOffset = 0

	case "G", "end":
		m.detailOffset = len(m.detailLines) - m.detailHeight()
		m.clampDetailScroll()

	case "ctrl+d":
		m.detailOffset += m.detailHeight() / 2
		m.clampDetailScroll()

	case "ctrl+u":
		m.detailOffset -= m.detailHeight() / 2
		m.clampDetailScroll()

	case "c", "C":
		md := strings.TrimRight(strings.Join(m.detailLines, "\n"), "\n")
		if msg.String() == "c" {
			_, body := things.SplitFrontmatter(md)
			md = strings.TrimLeft(body, "\n")
		}
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(md)
		if err := cmd.Run(); err != nil {
			m.err = err
		} else {
			m.status = "Copied to clipboard"
			return m, clearStatusAfter(2 * time.Second)
		}

	case "x":
		if m.cursor < len(m.items) {
			todo := m.items[m.cursor]
			switch todo.Status {
			case things.StatusCompleted:
				b := things.UpdateTodo(todo.ID, m.authToken, things.UpdateTodoOptions{Status: things.StatusPtr(things.StatusOpen)})
				_ = b.Open()
			case things.StatusOpen:
				b := things.UpdateTodo(todo.ID, m.authToken, things.UpdateTodoOptions{Status: things.StatusPtr(things.StatusCompleted)})
				if err := b.Open(); err != nil {
					m.err = err
				}
			}
		}

	case "ctrl+x":
		if m.cursor < len(m.items) {
			todo := m.items[m.cursor]
			switch todo.Status {
			case things.StatusCanceled:
				b := things.UpdateTodo(todo.ID, m.authToken, things.UpdateTodoOptions{Status: things.StatusPtr(things.StatusOpen)})
				_ = b.Open()
			case things.StatusOpen:
				b := things.UpdateTodo(todo.ID, m.authToken, things.UpdateTodoOptions{Status: things.StatusPtr(things.StatusCanceled)})
				if err := b.Open(); err != nil {
					m.err = err
				}
			}
		}

	case "X":
		if err := things.LogCompleted(); err != nil {
			m.err = err
		}

	case "e":
		if m.cursor < len(m.items) {
			return m.openEditor()
		}

	case "o":
		if m.cursor < len(m.items) {
			todo := m.items[m.cursor]
			if err := things.ShowItem(todo.ID).OpenForeground(); err != nil {
				m.err = err
			}
		}

	default:
		if m.cursor < len(m.items) {
			for _, action := range m.actions {
				if msg.String() == action.Key {
					return m.runAction(action)
				}
			}
		}
	}

	return m, nil
}

func (m model) openEditor() (tea.Model, tea.Cmd) {
	todo, err := m.db.GetTodo(m.items[m.cursor].ID)
	if err != nil {
		m.err = err
		return m, nil
	}
	tempPath, err := editor.TempFile(things.TodoToMarkdown(todo))
	if err != nil {
		m.err = err
		return m, nil
	}
	c := exec.Command(editor.Command(), tempPath)
	return m, tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{todoID: todo.ID, tempPath: tempPath, err: err}
	})
}

func (m model) refresh() model {
	todos, err := m.loader()
	if err != nil {
		m.err = err
		return m
	}

	m.items = todos
	if m.cursor >= len(m.items) && len(m.items) > 0 {
		m.cursor = len(m.items) - 1
	}
	m.clampScroll()

	if m.detailOpen && m.cursor < len(m.items) {
		todo, err := m.db.GetTodo(m.items[m.cursor].ID)
		if err != nil {
			m.detailOpen = false
		} else {
			m.setDetailContent(things.TodoToMarkdown(todo))
			m.clampDetailScroll()
		}
	} else if m.detailOpen {
		m.detailOpen = false
	}

	return m
}

func (m model) View() string {
	if m.detailOpen {
		return m.viewDetail()
	}

	var b strings.Builder

	b.WriteString(fmt.Sprintf(" %s\n\n", strings.Title(m.title)))

	if len(m.items) == 0 && !m.adding {
		b.WriteString(dimStyle.Render("  No items."))
		b.WriteString("\n")
	}

	lh := m.listHeight()
	end := m.offset + lh
	if end > len(m.items) {
		end = len(m.items)
	}

	for i := m.offset; i < end; i++ {
		todo := m.items[i]

		if todo.StartBucket == things.StartBucketEvening {
			if i == 0 || m.items[i-1].StartBucket != things.StartBucketEvening {
				b.WriteString("\n")
				b.WriteString(boldStyle.Render("  This Evening"))
				b.WriteString("\n\n")
			}
		}

		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("> ")
		}

		rawName := todo.Name

		// cursor(2) + statusPrefix(2) = 4 fixed chars
		if m.width > 0 {
			avail := m.width - 4
			nameLen := len([]rune(rawName))
			if nameLen > avail && avail > 0 {
				runes := []rune(rawName)
				rawName = string(runes[:avail-1]) + "…"
			}
		}

		var name string
		switch todo.Status {
		case things.StatusCompleted:
			name = doneStyle.Render("✓ " + rawName)
		case things.StatusCanceled:
			name = cancelStyle.Render("✗ " + rawName)
		default:
			name = "  " + rawName
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, name))
	}

	if m.adding {
		b.WriteString(m.addInput.View())
		b.WriteString("\n")
	}

	if m.status != "" {
		b.WriteString("\n")
		b.WriteString(statusBarStyle.Render(m.status))
		b.WriteString("\n")
	}

	if m.err != nil {
		b.WriteString(fmt.Sprintf("\n  Error: %v\n", m.err))
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render(wrapHelp(m.listHelpSegments(), m.width)))

	return b.String()
}

func (m model) viewDetail() string {
	var b strings.Builder

	// Content
	dh := m.detailHeight()
	end := m.detailOffset + dh
	if end > len(m.detailLines) {
		end = len(m.detailLines)
	}

	inFrontmatter := false
	// Determine if we're inside frontmatter at the start of the visible window.
	for i := 0; i < m.detailOffset && i < len(m.detailLines); i++ {
		if strings.TrimSpace(m.detailLines[i]) == "---" {
			inFrontmatter = !inFrontmatter
		}
	}

	for i := m.detailOffset; i < end; i++ {
		line := m.detailLines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "---" {
			inFrontmatter = !inFrontmatter
			b.WriteString(dimStyle.Render(" "+line) + "\n")
			continue
		}

		if inFrontmatter {
			b.WriteString(dimStyle.Render(" "+line) + "\n")
		} else if strings.HasPrefix(trimmed, "# ") {
			b.WriteString(" " + boldStyle.Render(line) + "\n")
		} else if strings.HasPrefix(trimmed, "- [x]") {
			b.WriteString(" " + doneStyle.Render(line) + "\n")
		} else {
			b.WriteString(" " + line + "\n")
		}
	}

	// Help bar
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(wrapHelp(m.detailHelpSegments(), m.width)))

	return b.String()
}

func (m *model) applyEdit(todoID, tempPath string) error {
	src, err := os.ReadFile(tempPath)
	if err != nil {
		return fmt.Errorf("read temp file: %w", err)
	}

	_, opts, checklist, err := things.ParseTodoMarkdownForUpdate(string(src))
	if err != nil {
		return err
	}

	// Don't send `when` if it hasn't changed — Things treats any explicit `when`
	// as a reschedule, which resets todayIndex and moves the item to the top.
	if todo, err := m.db.GetTodo(todoID); err == nil {
		if todo.ActivationDate != nil && opts.When == todo.ActivationDate.Format("2006-01-02") {
			opts.When = ""
		}
	}

	payload, err := things.BuildTodoUpdateJSON(todoID, opts, checklist)
	if err != nil {
		return err
	}
	return things.JSONCommand(payload, m.authToken, false).Open()
}

func (m model) runAction(action config.Action) (tea.Model, tea.Cmd) {
	todo := m.items[m.cursor]

	var content string
	switch action.Input {
	case config.ActionInputJSON:
		b, err := json.MarshalIndent(todo, "", "  ")
		if err != nil {
			m.err = err
			return m, nil
		}
		content = string(b)
	case config.ActionInputMarkdown:
		content = things.TodoToMarkdown(&todo)
	case config.ActionInputID:
		content = todo.ID
	}

	var args []string
	for _, a := range action.Args {
		if action.InputMode == config.ActionInputModeArg {
			args = append(args, strings.ReplaceAll(a, "$1", content))
		} else {
			args = append(args, a)
		}
	}
	c := exec.Command(action.Command, args...)

	if action.InputMode == config.ActionInputModeStdin {
		c.Stdin = strings.NewReader(content)
	}

	switch action.Mode {
	case config.ActionModeRun:
		out, err := c.CombinedOutput()
		if err != nil {
			m.err = err
		} else {
			msg := strings.TrimSpace(string(out))
			if msg == "" {
				msg = action.Label + ": done"
			}
			m.status = msg
			return m, clearStatusAfter(2 * time.Second)
		}
		return m, nil

	default: // exec
		return m, tea.ExecProcess(c, func(err error) tea.Msg {
			return actionFinishedMsg{err: err}
		})
	}
}

func watchDB() tea.Cmd {
	return func() tea.Msg {
		dbPath, err := things.FindDBPath()
		if err != nil {
			return nil
		}

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return nil
		}
		defer watcher.Close()

		dir := filepath.Dir(dbPath)
		if err := watcher.Add(dir); err != nil {
			return nil
		}

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return nil
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					time.Sleep(100 * time.Millisecond)
					for {
						select {
						case <-watcher.Events:
						default:
							return dbChangedMsg{}
						}
					}
				}
			case _, ok := <-watcher.Errors:
				if !ok {
					return nil
				}
			}
		}
	}
}

func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return statusClearMsg{}
	})
}

// wrapHelp joins help segments with two-space separators, wrapping to new
// lines when a segment would exceed width. width <= 0 means no wrapping.
// setDetailContent stores the raw markdown and derives word-wrapped detail lines.
func (m *model) setDetailContent(markdown string) {
	m.detailMarkdown = markdown
	m.detailLines = wrapDetailContent(markdown, m.width)
}

// wrapDetailContent splits markdown into display lines, word-wrapping prose
// to min(width-1, 79) columns. Frontmatter delimiters and headings are not wrapped.
func wrapDetailContent(markdown string, width int) []string {
	maxWidth := 79
	if width > 1 && width-1 < maxWidth {
		maxWidth = width - 1
	}
	var result []string
	for _, line := range strings.Split(markdown, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == "---" || strings.HasPrefix(trimmed, "# ") {
			result = append(result, line)
			continue
		}
		result = append(result, wordWrapLine(line, maxWidth)...)
	}
	return result
}

// wordWrapLine breaks a single line into multiple lines at word boundaries.
func wordWrapLine(line string, width int) []string {
	if len([]rune(line)) <= width {
		return []string{line}
	}
	// Preserve leading indentation on continuation lines.
	indent := ""
	for _, r := range line {
		if r == ' ' || r == '\t' {
			indent += string(r)
		} else {
			break
		}
	}
	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{line}
	}
	var lines []string
	current := indent
	for _, word := range words {
		if current == indent {
			current += word
		} else if len([]rune(current))+1+len([]rune(word)) <= width {
			current += " " + word
		} else {
			lines = append(lines, current)
			current = indent + word
		}
	}
	if current != indent {
		lines = append(lines, current)
	}
	return lines
}

func wrapHelp(segments []string, width int) string {
	if width <= 0 {
		return strings.Join(segments, "  ")
	}
	var lines []string
	current := ""
	for _, seg := range segments {
		if current == "" {
			current = seg
		} else if len(current)+2+len(seg) <= width {
			current += "  " + seg
		} else {
			lines = append(lines, current)
			current = seg
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return strings.Join(lines, "\n")
}

func viewSupportsAdd(view string) bool {
	switch view {
	case "inbox", "today", "anytime", "someday":
		return true
	}
	return false
}

func (m model) canAdd() bool {
	return viewSupportsAdd(m.view) || m.addList != ""
}

func viewToWhen(view string) string {
	switch view {
	case "today":
		return "today"
	case "anytime":
		return "anytime"
	case "someday":
		return "someday"
	}
	return ""
}
