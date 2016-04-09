package darwin

import (
	"database/sql/driver"
	"strings"
	"testing"
	"time"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestMigrate(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Errorf("sqlmock.New().error != nil, wants nil")
	}

	defer db.Close()

	dialect := MySQLDialect{}

	mock.ExpectExec(escapeQuery(dialect.CreateTableSQL())).WillReturnResult(sqlmock.NewResult(0, 0))

	query := "CREATE TABLE people (id INT AUTO_INCREMENT NOT NULL, PRIMARY KEY (id));"
	migration := Migration{
		Version:     1.0,
		Description: "Creating table people",
		Script:      strings.NewReader(query),
	}

	mock.ExpectExec(escapeQuery(query)).WillReturnResult(sqlmock.NewResult(1, 1))

	migrations := []Migration{migration}

	mock.ExpectExec(escapeQuery(dialect.MigrateSQL())).WithArgs(
		1.0, "Creating table people", "7ebca1c6f05333a728a8db4629e8d543",
		anyRFC3339{}, sqlmock.AnyArg(), true).WillReturnResult(sqlmock.NewResult(1, 1))

	Migrate(db, dialect, migrations)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func escapeQuery(s string) string {
	s1 := strings.Replace(s, ")", "\\)", -1)
	s1 = strings.Replace(s1, "(", "\\(", -1)
	s1 = strings.Replace(s1, "?", "\\?", -1)
	return s1
}

type anyRFC3339 struct{}

// Match satisfies sqlmock.Argument interface
func (a anyRFC3339) Match(v driver.Value) bool {
	_, ok := v.(string)

	if !ok {
		return false
	}

	_, err := time.Parse(time.RFC3339, v.(string))

	if err != nil {
		return false
	}

	return true
}
