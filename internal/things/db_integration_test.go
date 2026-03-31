//go:build integration

package things

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// These tests compare DB reads against JXA reads to ensure the DB queries
// produce the same results as the Things GUI. JXA is the ground truth.
//
// Run with: go test -tags=integration -v ./internal/things/

func TestMain(m *testing.M) {
	// Quick check that Things is running and JXA works
	_, err := RunJXA(`Application("Things3").name()`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "SKIP: Things 3 not available: %v\n", err)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestTodayMatchesGUI(t *testing.T) {
	compareView(t, "Today", func(db *DB) ([]Todo, error) {
		return db.ListTodos("today")
	})
}

func TestInboxMatchesGUI(t *testing.T) {
	compareView(t, "Inbox", func(db *DB) ([]Todo, error) {
		return db.ListTodos("inbox")
	})
}

func TestUpcomingMatchesGUI(t *testing.T) {
	compareView(t, "Upcoming", func(db *DB) ([]Todo, error) {
		return db.ListTodos("upcoming")
	})
}

func TestAnytimeMatchesGUI(t *testing.T) {
	compareView(t, "Anytime", func(db *DB) ([]Todo, error) {
		return db.ListTodos("anytime")
	})
}

func TestSomedayMatchesGUI(t *testing.T) {
	compareView(t, "Someday", func(db *DB) ([]Todo, error) {
		return db.ListTodos("someday")
	})
}

func TestLogbookMatchesGUI(t *testing.T) {
	// Logbook can be very large; JXA times out enumerating all items.
	// Verify DB returns completed/canceled items sorted by stopDate DESC.
	db, err := OpenDB()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	dbTodos, err := db.ListTodos("logbook")
	if err != nil {
		t.Fatalf("db logbook: %v", err)
	}

	if len(dbTodos) == 0 {
		t.Skip("Logbook is empty")
	}

	// Verify all items are completed or canceled
	for _, todo := range dbTodos {
		if todo.Status != StatusCompleted && todo.Status != StatusCanceled {
			t.Errorf("logbook item %q (%s) has status %s, expected completed or canceled",
				todo.Name, todo.ID, todo.Status)
		}
	}

	// Verify sorted by completion/cancellation date descending
	for i := 1; i < len(dbTodos); i++ {
		prev := completionTime(dbTodos[i-1])
		curr := completionTime(dbTodos[i])
		if prev != nil && curr != nil && curr.After(*prev) {
			t.Errorf("logbook not sorted: item %d (%s) is newer than item %d (%s)",
				i, dbTodos[i].Name, i-1, dbTodos[i-1].Name)
			break
		}
	}

	t.Logf("logbook: %d items, all completed/canceled, sorted by date", len(dbTodos))
}

func completionTime(todo Todo) *time.Time {
	if todo.CompletionDate != nil {
		return todo.CompletionDate
	}
	return todo.CancellationDate
}

func TestProjectListMatchesGUI(t *testing.T) {
	jxaRaw, err := RunJXA(listProjectsScript)
	if err != nil {
		t.Fatalf("jxa projects: %v", err)
	}
	var jxaProjects []Project
	if err := json.Unmarshal(jxaRaw, &jxaProjects); err != nil {
		t.Fatalf("parse jxa projects: %v", err)
	}

	db, err := OpenDB()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	dbProjects, err := db.ListProjects()
	if err != nil {
		t.Fatalf("db projects: %v", err)
	}

	if len(dbProjects) != len(jxaProjects) {
		t.Errorf("count mismatch: db=%d jxa=%d", len(dbProjects), len(jxaProjects))
	}

	jxaByID := make(map[string]Project, len(jxaProjects))
	for _, p := range jxaProjects {
		jxaByID[p.ID] = p
	}

	for _, dbP := range dbProjects {
		jxaP, ok := jxaByID[dbP.ID]
		if !ok {
			t.Errorf("DB project %q (%s) not in JXA", dbP.Name, dbP.ID)
			continue
		}
		if dbP.Name != jxaP.Name {
			t.Errorf("project %s name: db=%q jxa=%q", dbP.ID, dbP.Name, jxaP.Name)
		}
	}
}

func TestTrashMatchesGUI(t *testing.T) {
	// DB caps trash at 50; verify the DB items are all present in JXA's Trash.
	jxaIDs := jxaListIDs(t, "Trash")
	jxaSet := make(map[string]bool, len(jxaIDs))
	for _, id := range jxaIDs {
		jxaSet[id] = true
	}

	db, err := OpenDB()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	dbTodos, err := db.ListTodos("trash")
	if err != nil {
		t.Fatalf("db trash: %v", err)
	}
	if len(dbTodos) == 0 {
		t.Skip("Trash is empty")
	}

	for _, todo := range dbTodos {
		if !jxaSet[todo.ID] {
			t.Errorf("DB trash item %q (%s) not found in JXA Trash", todo.Name, todo.ID)
		}
	}
	t.Logf("trash: %d DB items (limit 50) out of %d JXA items", len(dbTodos), len(jxaIDs))
}

func TestTomorrowStructure(t *testing.T) {
	// JXA has no "Tomorrow" named list; verify DB results structurally.
	db, err := OpenDB()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	todos, err := db.ListTodos("tomorrow")
	if err != nil {
		t.Fatalf("db tomorrow: %v", err)
	}
	if len(todos) == 0 {
		t.Skip("no tomorrow todos")
	}

	tomorrowDate := encodeThingsDate(time.Now().AddDate(0, 0, 1))
	for _, todo := range todos {
		if todo.Status != StatusOpen {
			t.Errorf("tomorrow todo %q has status %s, want open", todo.Name, todo.Status)
		}
		if todo.ActivationDate == nil {
			t.Errorf("tomorrow todo %q has no activation date", todo.Name)
			continue
		}
		if got := encodeThingsDate(*todo.ActivationDate); got != tomorrowDate {
			t.Errorf("tomorrow todo %q has date %s, want tomorrow",
				todo.Name, todo.ActivationDate.Format("2006-01-02"))
		}
	}
	t.Logf("tomorrow: %d todos", len(todos))
}

func TestTodayMorningSubsetOfToday(t *testing.T) {
	// JXA has no startBucket concept; verify morning items appear in today and have bucket=0.
	db, err := OpenDB()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	morning, err := db.ListTodos("today-morning")
	if err != nil {
		t.Fatalf("db today-morning: %v", err)
	}
	if len(morning) == 0 {
		t.Skip("no today-morning todos")
	}

	allToday, err := db.ListTodos("today")
	if err != nil {
		t.Fatalf("db today: %v", err)
	}
	todaySet := make(map[string]bool, len(allToday))
	for _, td := range allToday {
		todaySet[td.ID] = true
	}

	for _, todo := range morning {
		if !todaySet[todo.ID] {
			t.Errorf("today-morning todo %q (%s) not in today", todo.Name, todo.ID)
		}
		if todo.StartBucket != 0 {
			t.Errorf("today-morning todo %q has startBucket %d, want 0", todo.Name, todo.StartBucket)
		}
	}
	t.Logf("today-morning: %d todos", len(morning))
}

func TestTodayEveningSubsetOfToday(t *testing.T) {
	// JXA has no startBucket concept; verify evening items appear in today and have bucket=1.
	db, err := OpenDB()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	evening, err := db.ListTodos("today-evening")
	if err != nil {
		t.Fatalf("db today-evening: %v", err)
	}
	if len(evening) == 0 {
		t.Skip("no today-evening todos")
	}

	allToday, err := db.ListTodos("today")
	if err != nil {
		t.Fatalf("db today: %v", err)
	}
	todaySet := make(map[string]bool, len(allToday))
	for _, td := range allToday {
		todaySet[td.ID] = true
	}

	for _, todo := range evening {
		if !todaySet[todo.ID] {
			t.Errorf("today-evening todo %q (%s) not in today", todo.Name, todo.ID)
		}
		if todo.StartBucket != 1 {
			t.Errorf("today-evening todo %q has startBucket %d, want 1", todo.Name, todo.StartBucket)
		}
	}
	t.Logf("today-evening: %d todos", len(evening))
}

func TestGetTodoMatchesGUI(t *testing.T) {
	db, err := OpenDB()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Find a stable todo to compare — inbox first, fall back to today.
	todos, err := db.ListTodos("inbox")
	if err != nil || len(todos) == 0 {
		todos, err = db.ListTodos("today")
		if err != nil || len(todos) == 0 {
			t.Skip("no todos available to test GetTodo")
		}
	}

	id := todos[0].ID
	dbTodo, err := db.GetTodo(id)
	if err != nil {
		t.Fatalf("db GetTodo %s: %v", id, err)
	}
	jxaTodo, err := GetTodo(id)
	if err != nil {
		t.Fatalf("jxa GetTodo %s: %v", id, err)
	}

	if dbTodo.Name != jxaTodo.Name {
		t.Errorf("name: db=%q jxa=%q", dbTodo.Name, jxaTodo.Name)
	}
	if dbTodo.Status != jxaTodo.Status {
		t.Errorf("status: db=%q jxa=%q", dbTodo.Status, jxaTodo.Status)
	}
	if dbTodo.Notes != jxaTodo.Notes {
		t.Errorf("notes: db=%q jxa=%q", dbTodo.Notes, jxaTodo.Notes)
	}
	if dbTodo.TagNames != jxaTodo.TagNames {
		t.Errorf("tagNames: db=%q jxa=%q", dbTodo.TagNames, jxaTodo.TagNames)
	}
	t.Logf("GetTodo %s: %q (status=%s)", id, dbTodo.Name, dbTodo.Status)
}

func TestGetProjectMatchesGUI(t *testing.T) {
	db, err := OpenDB()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	projects, err := db.ListProjects()
	if err != nil {
		t.Fatalf("db ListProjects: %v", err)
	}
	if len(projects) == 0 {
		t.Skip("no open projects")
	}

	id := projects[0].ID
	dbProject, err := db.GetProject(id)
	if err != nil {
		t.Fatalf("db GetProject %s: %v", id, err)
	}
	jxaProject, err := GetProject(id)
	if err != nil {
		t.Fatalf("jxa GetProject %s: %v", id, err)
	}

	if dbProject.Name != jxaProject.Name {
		t.Errorf("name: db=%q jxa=%q", dbProject.Name, jxaProject.Name)
	}
	if dbProject.Status != jxaProject.Status {
		t.Errorf("status: db=%q jxa=%q", dbProject.Status, jxaProject.Status)
	}
	if dbProject.Notes != jxaProject.Notes {
		t.Errorf("notes: db=%q jxa=%q", dbProject.Notes, jxaProject.Notes)
	}
	t.Logf("GetProject %s: %q", id, dbProject.Name)
}

func TestProjectTodosMatchGUI(t *testing.T) {
	db, err := OpenDB()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	projects, err := db.ListProjects()
	if err != nil {
		t.Fatalf("db ListProjects: %v", err)
	}

	// Find first project that has todos.
	var name string
	var dbTodos []Todo
	for _, p := range projects {
		todos, err := db.ListProjectTodos(p.Name)
		if err != nil || len(todos) == 0 {
			continue
		}
		name, dbTodos = p.Name, todos
		break
	}
	if name == "" {
		t.Skip("no projects with open todos")
	}

	jxaTodos, err := ListProjectTodos(name)
	if err != nil {
		t.Fatalf("jxa ListProjectTodos %q: %v", name, err)
	}

	dbIDs := make([]string, len(dbTodos))
	for i, td := range dbTodos {
		dbIDs[i] = td.ID
	}
	jxaIDs := make([]string, len(jxaTodos))
	for i, td := range jxaTodos {
		jxaIDs[i] = td.ID
	}

	onlyJXA, onlyDB, _ := setDiff(jxaIDs, dbIDs)
	if len(onlyJXA) > 0 || len(onlyDB) > 0 {
		t.Errorf("project %q todo mismatch: db=%d jxa=%d", name, len(dbIDs), len(jxaIDs))
		for _, id := range onlyJXA {
			t.Logf("  in JXA only: %s", id)
		}
		for _, id := range onlyDB {
			t.Logf("  in DB only: %s", id)
		}
	}
	t.Logf("project %q: %d todos", name, len(dbTodos))
}

func TestAreaTodosMatchGUI(t *testing.T) {
	db, err := OpenDB()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	areas, err := db.ListAreas()
	if err != nil {
		t.Fatalf("db ListAreas: %v", err)
	}

	// Find first area that has todos.
	var name string
	var dbTodos []Todo
	for _, a := range areas {
		todos, err := db.ListAreaTodos(a.Name)
		if err != nil || len(todos) == 0 {
			continue
		}
		name, dbTodos = a.Name, todos
		break
	}
	if name == "" {
		t.Skip("no areas with open todos")
	}

	jxaTodos, err := ListAreaTodos(name)
	if err != nil {
		t.Fatalf("jxa ListAreaTodos %q: %v", name, err)
	}

	dbIDs := make([]string, len(dbTodos))
	for i, td := range dbTodos {
		dbIDs[i] = td.ID
	}
	jxaIDs := make([]string, len(jxaTodos))
	for i, td := range jxaTodos {
		jxaIDs[i] = td.ID
	}

	onlyJXA, onlyDB, _ := setDiff(jxaIDs, dbIDs)
	if len(onlyJXA) > 0 || len(onlyDB) > 0 {
		t.Errorf("area %q todo mismatch: db=%d jxa=%d", name, len(dbIDs), len(jxaIDs))
		for _, id := range onlyJXA {
			t.Logf("  in JXA only: %s", id)
		}
		for _, id := range onlyDB {
			t.Logf("  in DB only: %s", id)
		}
	}
	t.Logf("area %q: %d todos", name, len(dbTodos))
}

func TestTagTodosMatchGUI(t *testing.T) {
	db, err := OpenDB()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	tags, err := db.ListTags()
	if err != nil {
		t.Fatalf("db ListTags: %v", err)
	}

	// Find first tag that has open todos.
	var name string
	var dbTodos []Todo
	for _, tag := range tags {
		todos, err := db.ListTagTodos(tag.Name)
		if err != nil || len(todos) == 0 {
			continue
		}
		name, dbTodos = tag.Name, todos
		break
	}
	if name == "" {
		t.Skip("no tags with open todos")
	}

	jxaTodos, err := ListTagTodos(name)
	if err != nil {
		t.Fatalf("jxa ListTagTodos %q: %v", name, err)
	}

	dbIDs := make([]string, len(dbTodos))
	for i, td := range dbTodos {
		dbIDs[i] = td.ID
	}
	jxaIDs := make([]string, len(jxaTodos))
	for i, td := range jxaTodos {
		jxaIDs[i] = td.ID
	}

	onlyJXA, onlyDB, _ := setDiff(jxaIDs, dbIDs)
	if len(onlyJXA) > 0 || len(onlyDB) > 0 {
		t.Errorf("tag %q todo mismatch: db=%d jxa=%d", name, len(dbIDs), len(jxaIDs))
		for _, id := range onlyJXA {
			t.Logf("  in JXA only: %s", id)
		}
		for _, id := range onlyDB {
			t.Logf("  in DB only: %s", id)
		}
	}
	t.Logf("tag %q: %d todos", name, len(dbTodos))
}

func TestSearchTodosStructure(t *testing.T) {
	// No JXA search equivalent; verify structural properties of results.
	db, err := OpenDB()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Derive a search term from the first inbox todo title.
	todos, err := db.ListTodos("inbox")
	if err != nil || len(todos) == 0 {
		t.Skip("inbox empty, cannot derive search term")
	}
	term := strings.Fields(todos[0].Name)[0]

	results, err := db.SearchTodos(term)
	if err != nil {
		t.Fatalf("SearchTodos %q: %v", term, err)
	}

	termLower := strings.ToLower(term)
	for _, todo := range results {
		if todo.Status != StatusOpen {
			t.Errorf("search result %q has status %s, want open", todo.Name, todo.Status)
		}
		if !strings.Contains(strings.ToLower(todo.Name), termLower) &&
			!strings.Contains(strings.ToLower(todo.Notes), termLower) {
			t.Errorf("search result %q does not contain term %q in title or notes", todo.Name, term)
		}
	}
	t.Logf("search %q: %d results", term, len(results))
}

func TestTagListMatchesGUI(t *testing.T) {
	jxaRaw, err := RunJXA(listTagsScript)
	if err != nil {
		t.Fatalf("jxa tags: %v", err)
	}
	var jxaTags []Tag
	if err := json.Unmarshal(jxaRaw, &jxaTags); err != nil {
		t.Fatalf("parse jxa tags: %v", err)
	}

	db, err := OpenDB()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	dbTags, err := db.ListTags()
	if err != nil {
		t.Fatalf("db tags: %v", err)
	}

	if len(dbTags) != len(jxaTags) {
		t.Errorf("count mismatch: db=%d jxa=%d", len(dbTags), len(jxaTags))
	}

	jxaByID := make(map[string]Tag, len(jxaTags))
	for _, tg := range jxaTags {
		jxaByID[tg.ID] = tg
	}

	for _, dbT := range dbTags {
		jxaT, ok := jxaByID[dbT.ID]
		if !ok {
			t.Errorf("DB tag %q (%s) not in JXA", dbT.Name, dbT.ID)
			continue
		}
		if dbT.Name != jxaT.Name {
			t.Errorf("tag %s name: db=%q jxa=%q", dbT.ID, dbT.Name, jxaT.Name)
		}
	}
}
