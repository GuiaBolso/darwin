package darwin

import (
	"database/sql"
	"log"
	"testing"

	_ "modernc.org/ql/driver"
)

func TestQLDialect(t *testing.T) {
	migrations := []Migration{
		{
			Version:     1,
			Description: "Creating table posts",
			Script: `CREATE TABLE posts (
						id int,
						title 	string,
					 );;`,
		},
		{
			Version:     2,
			Description: "Adding column body",
			Script:      "ALTER TABLE posts ADD body string;",
		},
	}
	db, err := sql.Open("ql-mem", "test.db")
	if err != nil {
		t.Fatal(err)
	}

	dv, err := NewGenericDriver(db, QLDialect{})
	if err != nil {
		t.Errorf("unable to construct driver: %s", err)
	}

	d := New(dv, migrations)
	err = d.Migrate()
	if err != nil {
		t.Fatal(err)
	}
	if !hasTable(db, "posts", t) {
		t.Error("expected the tble posts to exist")
	}
	cols := getAllColumns(db, "posts", t)
	if len(cols) != 3 {
		t.Errorf("expected 3 columns got %d", len(cols))
	}
}

func hasTable(db *sql.DB, tableName string, t *testing.T) bool {
	querry := "select count() from __Table where Name=$1"
	var count int
	err := db.QueryRow(querry, tableName).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	return count > 0
}

func getAllColumns(db *sql.DB, tableName string, t *testing.T) []string {
	var result []string
	query := `select Name from __Column where TableName=$1`
	rows, err := db.Query(query, tableName)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatal(err)
		}
		result = append(result, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return result
}
