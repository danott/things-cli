package things

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const defaultDBDir = "Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac"

// DB provides read-only access to the Things SQLite database.
type DB struct {
	db *sql.DB
}

// OpenDB opens the Things database in read-only mode.
func OpenDB() (*DB, error) {
	path, err := FindDBPath()
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro", path))
	if err != nil {
		return nil, fmt.Errorf("open things db: %w", err)
	}
	// Verify we can read
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("things db ping: %w", err)
	}
	return &DB{db: db}, nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

func FindDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	containerDir := filepath.Join(home, defaultDBDir)

	// Use glob instead of ReadDir — ReadDir is ~1.5s on Group Containers, glob is ~2ms
	pattern := filepath.Join(containerDir, "ThingsData-*", "Things Database.thingsdatabase", "main.sqlite")
	matches, err := filepath.Glob(pattern)
	if err == nil && len(matches) > 0 {
		return matches[0], nil
	}

	// Fallback: old location (pre-2023)
	oldPath := filepath.Join(containerDir, "Things Database.thingsdatabase", "main.sqlite")
	if _, err := os.Stat(oldPath); err == nil {
		return oldPath, nil
	}
	return "", fmt.Errorf("Things database not found in %s", containerDir)
}

// Things date encoding: (year << 16) + (month << 12) + (day << 7)
func decodeThingsDate(encoded int64) time.Time {
	day := int((encoded >> 7) & 0x1F)
	month := time.Month((encoded >> 12) & 0xF)
	year := int(encoded >> 16)
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
}

func encodeThingsDate(t time.Time) int64 {
	return (int64(t.Year()) << 16) + (int64(t.Month()) << 12) + (int64(t.Day()) << 7)
}

func cocoaToTime(ts float64) time.Time {
	// Core Data timestamps: seconds since 2001-01-01 00:00:00 UTC
	epoch := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
	return epoch.Add(time.Duration(ts * float64(time.Second)))
}

func timeToCocoa(t time.Time) float64 {
	epoch := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
	return t.Sub(epoch).Seconds()
}

func optionalTime(ts *float64) *time.Time {
	if ts == nil {
		return nil
	}
	t := cocoaToTime(*ts)
	return &t
}

func optionalThingsDate(encoded *int64) *time.Time {
	if encoded == nil || *encoded == 0 {
		return nil
	}
	t := decodeThingsDate(*encoded)
	return &t
}

// Task type constants
const (
	taskTypeTodo    = 0
	taskTypeProject = 1
	taskTypeHeading = 2
)

// Task status constants
const (
	taskStatusOpen      = 0
	taskStatusCanceled  = 2
	taskStatusCompleted = 3
)

func dbStatusToStatus(s int) Status {
	switch s {
	case taskStatusCompleted:
		return StatusCompleted
	case taskStatusCanceled:
		return StatusCanceled
	default:
		return StatusOpen
	}
}

const todosQuery = `
SELECT t.uuid, t.title, t.status, t.notes,
       t.startDate, t.deadline, t.creationDate, t.userModificationDate, t.stopDate,
       t.todayIndex, t.startBucket,
       p.title, p.uuid,
       a.title, a.uuid,
       GROUP_CONCAT(tag.title)
FROM TMTask t
LEFT JOIN TMTask p ON t.project = p.uuid
LEFT JOIN TMArea a ON t.area = a.uuid
LEFT JOIN TMTaskTag tt ON t.uuid = tt.tasks
LEFT JOIN TMTag tag ON tt.tags = tag.uuid
WHERE t.type = 0 AND t.trashed = 0
`

func (d *DB) scanTodos(rows *sql.Rows) ([]Todo, error) {
	var todos []Todo
	for rows.Next() {
		var (
			t            Todo
			status       int
			startDate    *int64
			deadline     *int64
			creationDate *float64
			modDate      *float64
			stopDate     *float64
			todayIndex   *int64
			startBucket  *int64
			projectName  *string
			projectID    *string
			areaName     *string
			areaID       *string
			tags         *string
		)
		err := rows.Scan(&t.ID, &t.Name, &status, &t.Notes,
			&startDate, &deadline, &creationDate, &modDate, &stopDate,
			&todayIndex, &startBucket, &projectName, &projectID, &areaName, &areaID, &tags)
		if err != nil {
			return nil, fmt.Errorf("scan todo: %w", err)
		}
		t.Status = dbStatusToStatus(status)
		t.ActivationDate = optionalThingsDate(startDate)
		t.DueDate = optionalThingsDate(deadline)
		t.CreationDate = optionalTime(creationDate)
		t.ModificationDate = optionalTime(modDate)
		if stopDate != nil {
			st := cocoaToTime(*stopDate)
			if t.Status == StatusCompleted {
				t.CompletionDate = &st
			} else if t.Status == StatusCanceled {
				t.CancellationDate = &st
			}
		}
		if startBucket != nil {
			t.StartBucket = int(*startBucket)
		}
		if projectName != nil {
			t.ProjectName = *projectName
		}
		if projectID != nil {
			t.ProjectID = *projectID
		}
		if areaName != nil {
			t.AreaName = *areaName
		}
		if areaID != nil {
			t.AreaID = *areaID
		}
		if tags != nil {
			t.TagNames = *tags
		}
		todos = append(todos, t)
	}
	return todos, rows.Err()
}

// DBListTodos returns todos from a named view.
func (d *DB) ListTodos(view string) ([]Todo, error) {
	today := encodeThingsDate(time.Now())

	var query string
	var args []interface{}

	switch strings.ToLower(view) {
	case "inbox":
		query = todosQuery + " AND t.start = 0 AND t.status = 0 GROUP BY t.uuid ORDER BY t.\"index\""
	case "today":
		query = todosQuery + " AND t.status = 0 AND t.startDate IS NOT NULL AND t.startDate <= ? GROUP BY t.uuid ORDER BY t.startBucket ASC, t.startDate DESC, t.todayIndex ASC"
		args = append(args, today)
	case "today-morning":
		query = todosQuery + " AND t.status = 0 AND t.startDate IS NOT NULL AND t.startDate <= ? AND t.startBucket = 0 GROUP BY t.uuid ORDER BY t.startDate DESC, t.todayIndex ASC"
		args = append(args, today)
	case "today-evening":
		query = todosQuery + " AND t.status = 0 AND t.startDate IS NOT NULL AND t.startDate <= ? AND t.startBucket = 1 GROUP BY t.uuid ORDER BY t.startDate DESC, t.todayIndex ASC"
		args = append(args, today)
	case "upcoming":
		query = todosQuery + ` AND t.status = 0
			AND (
				(t.startDate IS NOT NULL AND t.startDate > ?)
				OR (t.todayIndexReferenceDate IS NOT NULL AND t.todayIndexReferenceDate > ?)
				OR (t.rt1_instanceCreationStartDate IS NOT NULL AND t.rt1_instanceCreationStartDate > ?)
			)
			GROUP BY t.uuid
			ORDER BY COALESCE(t.startDate, t.todayIndexReferenceDate) IS NULL,
			         COALESCE(t.startDate, t.todayIndexReferenceDate, t.rt1_instanceCreationStartDate),
			         t.todayIndex ASC`
		args = append(args, today, today, today)
	case "anytime":
		query = todosQuery + " AND t.status = 0 AND t.start = 1 GROUP BY t.uuid ORDER BY t.\"index\""
	case "someday":
		query = todosQuery + ` AND t.status = 0 AND t.start = 2 AND t.startDate IS NULL
			AND t.project IS NULL
			AND (t.todayIndexReferenceDate IS NULL OR t.todayIndexReferenceDate <= ?)
			AND (t.rt1_instanceCreationStartDate IS NULL OR t.rt1_instanceCreationStartDate <= ?)
			GROUP BY t.uuid ORDER BY t."index"`
		args = append(args, today, today)
	case "logbook":
		query = todosQuery + " AND t.status IN (2, 3) GROUP BY t.uuid ORDER BY t.stopDate DESC LIMIT 50"
	case "trash":
		query = `SELECT t.uuid, t.title, t.status, t.notes,
			t.startDate, t.deadline, t.creationDate, t.userModificationDate, t.stopDate,
			t.todayIndex, t.startBucket, p.title, p.uuid, a.title, a.uuid, GROUP_CONCAT(tag.title)
			FROM TMTask t
			LEFT JOIN TMTask p ON t.project = p.uuid
			LEFT JOIN TMArea a ON t.area = a.uuid
			LEFT JOIN TMTaskTag tt ON t.uuid = tt.tasks
			LEFT JOIN TMTag tag ON tt.tags = tag.uuid
			WHERE t.type = 0 AND t.trashed = 1
			GROUP BY t.uuid ORDER BY t.userModificationDate DESC LIMIT 50`
	case "tomorrow":
		tomorrow := encodeThingsDate(time.Now().AddDate(0, 0, 1))
		query = todosQuery + " AND t.status = 0 AND t.startDate = ? GROUP BY t.uuid ORDER BY t.\"index\""
		args = append(args, tomorrow)
	default:
		return nil, fmt.Errorf("unknown view: %s", view)
	}

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query %s: %w", view, err)
	}
	defer rows.Close()
	return d.scanTodos(rows)
}

// ListTodosWithCompleted returns todos from a named view, including items
// completed or canceled but not yet logged. Things tracks the log cutoff in
// TMSettings.manualLogDate — items with stopDate > manualLogDate are still
// visible in views; items at or before the cutoff have been "logged away."
func (d *DB) ListTodosWithCompleted(view string) ([]Todo, error) {
	today := encodeThingsDate(time.Now())

	// Subquery for the log cutoff timestamp.
	logCutoff := `(SELECT manualLogDate FROM TMSettings LIMIT 1)`

	var query string
	var args []interface{}

	switch strings.ToLower(view) {
	case "inbox":
		query = todosQuery + ` AND t.start = 0 AND (
				t.status = 0
				OR (t.status IN (2, 3) AND t.stopDate > ` + logCutoff + `)
			) GROUP BY t.uuid ORDER BY t.status ASC, t."index"`
	case "today":
		query = todosQuery + ` AND t.startDate IS NOT NULL AND t.startDate <= ? AND (
				t.status = 0
				OR (t.status IN (2, 3) AND t.stopDate > ` + logCutoff + `)
			) GROUP BY t.uuid ORDER BY t.status ASC, t.startBucket ASC, t.startDate DESC, t.todayIndex ASC`
		args = append(args, today)
	case "upcoming":
		query = todosQuery + ` AND (
				(t.status = 0 AND (
					(t.startDate IS NOT NULL AND t.startDate > ?)
					OR (t.todayIndexReferenceDate IS NOT NULL AND t.todayIndexReferenceDate > ?)
					OR (t.rt1_instanceCreationStartDate IS NOT NULL AND t.rt1_instanceCreationStartDate > ?)
				))
				OR (t.status IN (2, 3) AND t.stopDate > ` + logCutoff + `
					AND t.startDate IS NOT NULL AND t.startDate > ?)
			) GROUP BY t.uuid
			ORDER BY t.status ASC,
			         COALESCE(t.startDate, t.todayIndexReferenceDate) IS NULL,
			         COALESCE(t.startDate, t.todayIndexReferenceDate, t.rt1_instanceCreationStartDate),
			         t.todayIndex ASC`
		args = append(args, today, today, today, today)
	case "anytime":
		query = todosQuery + ` AND t.start = 1 AND (
				t.status = 0
				OR (t.status IN (2, 3) AND t.stopDate > ` + logCutoff + `)
			) GROUP BY t.uuid ORDER BY t.status ASC, t."index"`
	case "someday":
		query = todosQuery + ` AND t.start = 2 AND t.startDate IS NULL AND t.project IS NULL
			AND (t.todayIndexReferenceDate IS NULL OR t.todayIndexReferenceDate <= ?)
			AND (t.rt1_instanceCreationStartDate IS NULL OR t.rt1_instanceCreationStartDate <= ?)
			AND (
				t.status = 0
				OR (t.status IN (2, 3) AND t.stopDate > ` + logCutoff + `)
			) GROUP BY t.uuid ORDER BY t.status ASC, t."index"`
		args = append(args, today, today)
	default:
		return d.ListTodos(view)
	}

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query %s (interactive): %w", view, err)
	}
	defer rows.Close()
	return d.scanTodos(rows)
}

// ListProjectTodos returns todos belonging to a project, matched by UUID or title.
func (d *DB) ListProjectTodos(titleOrID string) ([]Todo, error) {
	query := todosQuery + " AND (p.uuid = ? OR p.title = ?) AND t.status = 0 GROUP BY t.uuid ORDER BY t.\"index\""
	rows, err := d.db.Query(query, titleOrID, titleOrID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return d.scanTodos(rows)
}

// ListAreaTodos returns todos belonging to an area, matched by UUID or title.
func (d *DB) ListAreaTodos(titleOrID string) ([]Todo, error) {
	query := todosQuery + " AND (a.uuid = ? OR a.title = ?) AND t.status = 0 GROUP BY t.uuid ORDER BY t.\"index\""
	rows, err := d.db.Query(query, titleOrID, titleOrID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return d.scanTodos(rows)
}

// ListTagTodos returns todos with a specific tag, matched by UUID or title.
func (d *DB) ListTagTodos(titleOrID string) ([]Todo, error) {
	query := todosQuery + " AND (tag.uuid = ? OR tag.title = ?) AND t.status = 0 GROUP BY t.uuid ORDER BY t.\"index\""
	rows, err := d.db.Query(query, titleOrID, titleOrID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return d.scanTodos(rows)
}

// GetTodo returns a single todo by UUID or title, including checklist items.
func (d *DB) GetTodo(titleOrID string) (*Todo, error) {
	query := todosQuery + " AND (t.uuid = ? OR t.title = ?) GROUP BY t.uuid"
	rows, err := d.db.Query(query, titleOrID, titleOrID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	todos, err := d.scanTodos(rows)
	if err != nil {
		return nil, err
	}
	if len(todos) == 0 {
		return nil, fmt.Errorf("todo not found: %s", titleOrID)
	}
	todo := &todos[0]
	items, err := d.ListChecklistItems(todo.ID)
	if err != nil {
		return nil, err
	}
	todo.ChecklistItems = items
	return todo, nil
}

// ListLogbook returns completed/canceled todos with pagination and date filtering.
func (d *DB) ListLogbook(opts LogbookOptions) ([]Todo, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}

	baseQuery := todosQuery + " AND t.status IN (2, 3)"
	var args []interface{}

	if opts.Since != nil {
		baseQuery += " AND t.stopDate >= ?"
		args = append(args, timeToCocoa(*opts.Since))
	}
	if opts.Until != nil {
		baseQuery += " AND t.stopDate <= ?"
		args = append(args, timeToCocoa(*opts.Until))
	}

	baseQuery += " GROUP BY t.uuid ORDER BY t.stopDate DESC LIMIT ? OFFSET ?"
	args = append(args, limit, opts.Offset)

	rows, err := d.db.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query logbook: %w", err)
	}
	defer rows.Close()
	return d.scanTodos(rows)
}

// ListProjects returns all open projects.
func (d *DB) ListProjects() ([]Project, error) {
	rows, err := d.db.Query(`
		SELECT t.uuid, t.title, t.status, COALESCE(t.notes, ''),
		       COALESCE(a.title, ''), t.startDate, t.deadline,
		       GROUP_CONCAT(tag.title)
		FROM TMTask t
		LEFT JOIN TMArea a ON t.area = a.uuid
		LEFT JOIN TMTaskTag tt ON t.uuid = tt.tasks
		LEFT JOIN TMTag tag ON tt.tags = tag.uuid
		WHERE t.type = 1 AND t.trashed = 0 AND t.status = 0
		GROUP BY t.uuid
		ORDER BY t."index"
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		var status int
		var startDate, deadline *int64
		var tags *string
		if err := rows.Scan(&p.ID, &p.Name, &status, &p.Notes, &p.AreaName, &startDate, &deadline, &tags); err != nil {
			return nil, err
		}
		p.Status = dbStatusToStatus(status)
		p.ActivationDate = optionalThingsDate(startDate)
		p.DueDate = optionalThingsDate(deadline)
		if tags != nil {
			p.TagNames = *tags
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

// GetProject returns a project by UUID or title.
func (d *DB) GetProject(titleOrID string) (*Project, error) {
	row := d.db.QueryRow(`
		SELECT t.uuid, t.title, t.status, COALESCE(t.notes, ''),
		       COALESCE(a.title, ''), t.startDate, t.deadline,
		       GROUP_CONCAT(tag.title)
		FROM TMTask t
		LEFT JOIN TMArea a ON t.area = a.uuid
		LEFT JOIN TMTaskTag tt ON t.uuid = tt.tasks
		LEFT JOIN TMTag tag ON tt.tags = tag.uuid
		WHERE (t.uuid = ? OR t.title = ?) AND t.type = 1
		GROUP BY t.uuid
	`, titleOrID, titleOrID)
	var p Project
	var status int
	var startDate, deadline *int64
	var tags *string
	if err := row.Scan(&p.ID, &p.Name, &status, &p.Notes, &p.AreaName, &startDate, &deadline, &tags); err != nil {
		return nil, fmt.Errorf("project not found: %s", titleOrID)
	}
	p.Status = dbStatusToStatus(status)
	p.ActivationDate = optionalThingsDate(startDate)
	p.DueDate = optionalThingsDate(deadline)
	if tags != nil {
		p.TagNames = *tags
	}
	return &p, nil
}

// ListAreas returns all areas ordered by index.
func (d *DB) ListAreas() ([]Area, error) {
	rows, err := d.db.Query(`SELECT uuid, title FROM TMArea ORDER BY "index"`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []Area
	for rows.Next() {
		var a Area
		if err := rows.Scan(&a.ID, &a.Name); err != nil {
			return nil, err
		}
		areas = append(areas, a)
	}
	return areas, rows.Err()
}

// GetArea returns an area by ID or name.
func (d *DB) GetArea(idOrName string) (*Area, error) {
	row := d.db.QueryRow(`SELECT uuid, title FROM TMArea WHERE uuid = ? OR title = ? LIMIT 1`, idOrName, idOrName)
	var a Area
	if err := row.Scan(&a.ID, &a.Name); err != nil {
		return nil, fmt.Errorf("area not found: %s", idOrName)
	}
	return &a, nil
}

// ListChecklistItems returns checklist items for a todo, ordered by index.
func (d *DB) ListChecklistItems(todoID string) ([]ChecklistItem, error) {
	rows, err := d.db.Query(`
		SELECT uuid, COALESCE(title, ''), status
		FROM TMChecklistItem
		WHERE task = ?
		ORDER BY "index"
	`, todoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ChecklistItem
	for rows.Next() {
		var item ChecklistItem
		var status int
		if err := rows.Scan(&item.ID, &item.Name, &status); err != nil {
			return nil, err
		}
		item.Status = dbStatusToStatus(status)
		items = append(items, item)
	}
	return items, rows.Err()
}

// SearchTodos returns open todos matching a query against title and notes.
func (d *DB) SearchTodos(query string) ([]Todo, error) {
	like := "%" + query + "%"
	q := todosQuery + " AND t.status = 0 AND (t.title LIKE ? OR t.notes LIKE ?) GROUP BY t.uuid ORDER BY t.userModificationDate DESC LIMIT 50"
	rows, err := d.db.Query(q, like, like)
	if err != nil {
		return nil, fmt.Errorf("search todos: %w", err)
	}
	defer rows.Close()
	return d.scanTodos(rows)
}

// ListTags returns all tags.
func (d *DB) ListTags() ([]Tag, error) {
	rows, err := d.db.Query(`
		SELECT t.uuid, t.title, COALESCE(p.title, '')
		FROM TMTag t
		LEFT JOIN TMTag p ON t.parent = p.uuid
		ORDER BY t."index"
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.ParentTag); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}
