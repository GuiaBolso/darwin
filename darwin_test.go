package darwin

import (
	"database/sql/driver"
	"fmt"
	"regexp"
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

	mock.ExpectBegin()
	mock.ExpectExec(escapeQuery(query)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	migrations := []Migration{migration}

	mock.ExpectExec(escapeQuery(dialect.MigrateSQL())).WithArgs(
		1.0, "Creating table people", "7ebca1c6f05333a728a8db4629e8d543",
		&anyRFC3339{}, sqlmock.AnyArg(), true).WillReturnResult(sqlmock.NewResult(1, 1))

	Migrate(db, dialect, migrations)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func escapeQuery(s string) string {
	s1 := strings.NewReplacer(
		")", "\\)",
		"(", "\\(",
		"?", "\\?",
		"\n", " ",
		"\r", " ",
		"\t", " ",
	).Replace(s)

	re := regexp.MustCompile("\\s+")
	s1 = strings.TrimSpace(re.ReplaceAllString(s1, " "))
	return s1
}

type anyRFC3339 struct {
	value interface{}
}

func (a *anyRFC3339) String() string {
	return fmt.Sprintf("%v", a.value)
}

// Match satisfies sqlmock.Argument interface
func (a *anyRFC3339) Match(v driver.Value) bool {
	a.value = v

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
