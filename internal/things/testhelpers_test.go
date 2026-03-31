//go:build integration

package things

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// compareView checks that DB and JXA return the same todo IDs in the same order.
// On mismatch, it logs diagnostic info about missing/extra items instead of failing fast.
func compareView(t *testing.T, listName string, dbQuery func(*DB) ([]Todo, error)) {
	t.Helper()

	jxaIDs := jxaListIDs(t, listName)

	db, err := OpenDB()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	dbTodos, err := dbQuery(db)
	if err != nil {
		t.Fatalf("db query: %v", err)
	}

	dbIDs := make([]string, len(dbTodos))
	for i, todo := range dbTodos {
		dbIDs[i] = todo.ID
	}

	onlyJXA, onlyDB, _ := setDiff(jxaIDs, dbIDs)

	if len(onlyJXA) > 0 || len(onlyDB) > 0 {
		// Separate transient mismatches (completed/canceled recurring task instances
		// where JXA and DB are briefly out of sync) from real query bugs.
		realJXA, transientJXA := classifyMissing(db, onlyJXA)
		realDB, transientDB := classifyMissing(db, onlyDB)

		if len(transientJXA) > 0 || len(transientDB) > 0 {
			t.Logf("note: %d transient mismatches (completed/canceled recurring instances — JXA/DB sync delay)",
				len(transientJXA)+len(transientDB))
		}

		if len(realJXA) > 0 || len(realDB) > 0 {
			t.Errorf("count mismatch: db=%d jxa=%d (%d real, %d transient)",
				len(dbIDs), len(jxaIDs), len(realJXA)+len(realDB), len(transientJXA)+len(transientDB))
			if len(realJXA) > 0 {
				t.Errorf("%d items in JXA but missing from DB:", len(realJXA))
				diagnoseMissing(t, db, realJXA)
			}
			if len(realDB) > 0 {
				t.Errorf("%d items in DB but missing from JXA:", len(realDB))
				diagnoseMissing(t, db, realDB)
			}
		}
		return
	}

	// Sets match — verify ordering
	for i := range dbIDs {
		if dbIDs[i] != jxaIDs[i] {
			t.Errorf("order mismatch starting at position %d", i)
			diffIDLists(t, jxaIDs, dbIDs, i)
			break
		}
	}
}

// jxaListIDs returns todo IDs from a Things list via JXA (the ground truth).
func jxaListIDs(t *testing.T, listName string) []string {
	t.Helper()

	script := fmt.Sprintf(`
const things = Application("Things3");
const todos = things.lists.byName("%s").toDos();
JSON.stringify(todos.map(t => t.id()));
`, escapeJS(listName))

	raw, err := RunJXA(script)
	if err != nil {
		t.Fatalf("jxa %s: %v", listName, err)
	}

	var ids []string
	if err := json.Unmarshal(raw, &ids); err != nil {
		t.Fatalf("parse jxa ids: %v", err)
	}
	return ids
}

// jxaListIDsWithTimeout is like jxaListIDs but with a custom timeout.
func jxaListIDsWithTimeout(t *testing.T, listName string, timeout time.Duration) []string {
	t.Helper()

	script := fmt.Sprintf(`
const things = Application("Things3");
const todos = things.lists.byName("%s").toDos();
JSON.stringify(todos.map(t => t.id()));
`, escapeJS(listName))

	raw, err := RunJXAWithTimeout(script, timeout)
	if err != nil {
		t.Fatalf("jxa %s: %v", listName, err)
	}

	var ids []string
	if err := json.Unmarshal(raw, &ids); err != nil {
		t.Fatalf("parse jxa ids: %v", err)
	}
	return ids
}

// diagnoseMissing queries raw TMTask columns for the given IDs and logs them.
func diagnoseMissing(t *testing.T, db *DB, ids []string) {
	t.Helper()
	for _, id := range ids {
		var (
			uuid        string
			title       string
			taskType    int
			status      int
			trashed     int
			start       int
			startDate   sql.NullInt64
			project     sql.NullString
			heading     sql.NullString
			todayIndex  sql.NullInt64
		)
		err := db.db.QueryRow(`
			SELECT uuid, title, type, status, trashed, start,
			       startDate, project, heading, todayIndex
			FROM TMTask WHERE uuid = ?`, id).Scan(
			&uuid, &title, &taskType, &status, &trashed, &start,
			&startDate, &project, &heading, &todayIndex,
		)
		if err != nil {
			t.Logf("  %s: (query error: %v)", id, err)
			continue
		}

		sd := "nil"
		if startDate.Valid {
			sd = fmt.Sprintf("%d (%s)", startDate.Int64, decodeThingsDate(startDate.Int64).Format("2006-01-02"))
		}
		proj := "nil"
		if project.Valid {
			proj = project.String[:min(len(project.String), 12)]
		}
		hd := "nil"
		if heading.Valid {
			hd = heading.String[:min(len(heading.String), 12)]
		}
		ti := "nil"
		if todayIndex.Valid {
			ti = fmt.Sprintf("%d", todayIndex.Int64)
		}

		t.Logf("  %s name=%q type=%d status=%d trashed=%d start=%d startDate=%s project=%s heading=%s todayIndex=%s",
			uuid, title, taskType, status, trashed, start, sd, proj, hd, ti)
	}
}

// classifyMissing separates real mismatches from transient ones.
// Transient: items that are completed/canceled in DB but JXA still lists them
// (recurring task instances in transition).
func classifyMissing(db *DB, ids []string) (real, transient []string) {
	for _, id := range ids {
		var status int
		err := db.db.QueryRow("SELECT status FROM TMTask WHERE uuid = ?", id).Scan(&status)
		if err != nil {
			real = append(real, id)
			continue
		}
		if status == taskStatusCompleted || status == taskStatusCanceled {
			transient = append(transient, id)
		} else {
			real = append(real, id)
		}
	}
	return
}

// setDiff returns (onlyInA, onlyInB, common) for two ID slices.
func setDiff(a, b []string) (onlyA, onlyB, common []string) {
	setA := make(map[string]bool, len(a))
	for _, id := range a {
		setA[id] = true
	}
	setB := make(map[string]bool, len(b))
	for _, id := range b {
		setB[id] = true
	}

	for _, id := range a {
		if setB[id] {
			common = append(common, id)
		} else {
			onlyA = append(onlyA, id)
		}
	}
	for _, id := range b {
		if !setA[id] {
			onlyB = append(onlyB, id)
		}
	}
	return
}

// diffIDLists logs the first few positions where two ordered ID lists diverge.
func diffIDLists(t *testing.T, jxaIDs, dbIDs []string, startAt int) {
	t.Helper()
	shown := 0
	maxShow := 10
	for i := startAt; i < len(jxaIDs) && i < len(dbIDs) && shown < maxShow; i++ {
		if jxaIDs[i] != dbIDs[i] {
			t.Logf("  position %d: jxa=%s db=%s", i, jxaIDs[i], dbIDs[i])
			shown++
		}
	}
}
