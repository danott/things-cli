package interactive

import (
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
	adding    bool
	addInput  textinput.Model
	authToken string
	err       error
	status    string

	// Detail view state
	detailOpen   bool
	detailLines  []string
	detailOffset int
}

type dbChangedMsg struct{}
type statusClearMsg struct{}
type editorFinishedMsg struct {
	todoID   string
	tempPath string
	err      error
}

func Run(db *things.DB, view, authToken string) error {
	loader := func() ([]things.Todo, error) {
		return db.ListTodosWithCompleted(view)
	}
	return RunWithLoader(db, view, view, loader, authToken)
}

// RunWithLoader starts the interactive TUI with a custom todo loader.
// title is shown as the list heading; view is used for add/when behavior.
func RunWithLoader(db *things.DB, title, view string, loader func() ([]things.Todo, error), authToken string) error {
	m, err := newModel(db, title, view, loader, authToken)
	if err != nil {
		return err
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func newModel(db *things.DB, title, view string, loader func() ([]things.Todo, error), authToken string) (model, error) {
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
		height:    24, // default, updated on WindowSizeMsg
	}, nil
}

func (m model) Init() tea.Cmd {
	return watchDB()
}

// listHeight returns how many item rows fit in the viewport.
// Reserve lines for: title (1) + blank (1) + help bar (2) + status (1) + add input (1)
func (m model) listHeight() int {
	reserved := 4 // title + blank + help + bottom padding
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

// detailHeight returns how many content lines fit in the detail viewport.
// Reserve: breadcrumb (1) + blank (1) + help bar (2).
func (m model) detailHeight() int {
	h := m.height - 4
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
		if m.detailOpen {
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
			m.detailLines = strings.Split(things.TodoToMarkdown(todo), "\n")
			m.detailOffset = 0
			m.detailOpen = true
		}

	case "e":
		if m.cursor < len(m.items) {
			return m.openEditor()
		}

	case "a":
		if viewSupportsAdd(m.view) {
			m.adding = true
			m.addInput.Reset()
			m.addInput.Focus()
			return m, textinput.Blink
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
		opts := things.AddTodoOptions{When: viewToWhen(m.view)}
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
	case "esc", "enter", "q":
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

	case "e":
		if m.cursor < len(m.items) {
			return m.openEditor()
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
			m.detailLines = strings.Split(things.TodoToMarkdown(todo), "\n")
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
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("> ")
		}

		name := todo.Name
		switch todo.Status {
		case things.StatusCompleted:
			name = doneStyle.Render("✓ " + name)
		case things.StatusCanceled:
			name = cancelStyle.Render("✗ " + name)
		default:
			name = "  " + name
		}

		extra := todoExtra(todo)
		if extra != "" {
			name += dimStyle.Render("  " + extra)
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

	help := "j/k: navigate  enter: view  e: edit  x: complete  ctrl+x: cancel  X: clear done"
	if viewSupportsAdd(m.view) {
		help += "  a: add"
	}
	help += "  q: quit"
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(help))

	return b.String()
}

func (m model) viewDetail() string {
	var b strings.Builder

	// Breadcrumb
	title := ""
	if m.cursor < len(m.items) {
		title = m.items[m.cursor].Name
	}
	b.WriteString(dimStyle.Render(" \u2039 ") + title + "\n\n")

	// Content
	dh := m.detailHeight()
	end := m.detailOffset + dh
	if end > len(m.detailLines) {
		end = len(m.detailLines)
	}

	inFrontmatter := false
	// Determine if we're inside frontmatter at the start of the visible window
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
	b.WriteString(dimStyle.Render("j/k: scroll  c: copy  e: edit  enter/esc: back"))

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

	payload, err := things.BuildTodoUpdateJSON(todoID, opts, checklist)
	if err != nil {
		return err
	}
	return things.JSONCommand(payload, m.authToken, false).Open()
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

func todoExtra(t things.Todo) string {
	var parts []string
	if t.ProjectName != "" {
		parts = append(parts, t.ProjectName)
	} else if t.AreaName != "" {
		parts = append(parts, t.AreaName)
	}
	if t.TagNames != "" {
		parts = append(parts, "["+t.TagNames+"]")
	}
	return strings.Join(parts, "  ")
}

func viewSupportsAdd(view string) bool {
	switch view {
	case "inbox", "today", "anytime", "someday":
		return true
	}
	return false
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
