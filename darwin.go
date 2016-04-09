package darwin

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"time"
)

// Migration represents a database migrations.
type Migration struct {
	Version     float64
	Description string
	Script      io.Reader
}

// Dialect is a interface used to abstract the different databases.
type Dialect interface {
	CreateTableSQL() string
	MigrateSQL() string
	LastVersionSQL() string
}

// Darwin is a helper struct to access the Validate and migratin functions
type Darwin struct{}

// Validate if the database migratins are applied and consistent
func (d Darwin) Validate() bool {
	return false
}

// Migrate executes the missing migrations in database
func (d Darwin) Migrate() error {
	return nil
}

// New returns a new Darwin struct
func New(db *sql.DB, d Dialect, migrations []Migration) Darwin {
	return Darwin{}
}

// NewForMySQL returns a new Darwin configured with MySQL dialect
func NewForMySQL(db *sql.DB, migrations []Migration) Darwin {
	return Darwin{}
}

// Validate if the database migratins are applied and consistent
func Validate(db *sql.DB, dialect Dialect, migrations []Migration) bool {
	return false
}

func createSchemaTable(db *sql.DB, dialect Dialect) error {
	_, err := db.Exec(dialect.CreateTableSQL())

	return err
}

// Migrate executes the missing migrations in database
func Migrate(db *sql.DB, dialect Dialect, migrations []Migration) error {
	err := createSchemaTable(db, dialect)

	sort.Sort(ByVersion(migrations))

	for _, migration := range migrations {
		script, err := ioutil.ReadAll(migration.Script)

		if err != nil {
			return err
		}

		start := time.Now()
		_, err = db.Exec(string(script))
		elapsed := time.Since(start)

		success := true

		if err != nil {
			success = false
		}

		_, err = db.Exec(dialect.MigrateSQL(),
			migration.Version,
			migration.Description,
			fmt.Sprintf("%x", md5.Sum(script)),
			time.Now().Format(time.RFC3339),
			elapsed.Seconds(),
			success,
		)

		if err != nil {
			return err
		}
	}

	return err
}

// MySQLDialect holds the definition of a MySQL dialect
type MySQLDialect struct{}

// CreateTableSQL returns a schema create table
func (m MySQLDialect) CreateTableSQL() string {
	return "CREATE TABLE IF NOT EXISTS darwin_migrations (id INT AUTO_INCREMENT, version FLOAT NOT NULL, description VARCHAR(255) NOT NULL, checksum VARCHAR(32) NOT NULL, applied_at DATETIME NOT NULL, execution_time FLOAT NOT NULL, success BOOL NOT NULL, PRIMARY KEY (id));"
}

// MigrateSQL returns a schema migrate table
func (m MySQLDialect) MigrateSQL() string {
	return "INSERT INTO darwin_migrations (version, description, checksum, applied_at, execution_time, success) VALUES (?, ?, ?, ?, ?, ?);"
}

// LastVersionSQL returns a new SQL fo get the last version in the database
func (m MySQLDialect) LastVersionSQL() string {
	return "SELECT version FROM darwin_migrations ORDER BY version DESC"
}

// ByVersion implements the Sort interface sorting bt Version
type ByVersion []Migration

func (b ByVersion) Len() int           { return len(b) }
func (b ByVersion) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByVersion) Less(i, j int) bool { return b[i].Version < b[j].Version }
