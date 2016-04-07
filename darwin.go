package darwin

import (
	"database/sql"
	"io"
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

// Migrate executes the missing migrations in database
func Migrate(db *sql.DB, dialect Dialect, migrations []Migration) error {
	return nil
}

// MySQLDialect holds the definition of a MySQL dialect
type MySQLDialect struct{}

// CreateTableSQL returns a schema create table
func (m MySQLDialect) CreateTableSQL() string {
	return "CREATE TABLE schema_migrations"
}

// MigrateSQL returns a schema migrate table
func (m MySQLDialect) MigrateSQL() string {
	return "INSERT INTO schema_migrations"
}

// LastVersionSQL returns a new SQL fo get the last version in the database
func (m MySQLDialect) LastVersionSQL() string {
	return "SELECT version FROM schema_migrations ORDER BY version DESC"
}

// ByVersion implements the Sort interface sorting bt Version
type ByVersion []Migration

func (b ByVersion) Len() int           { return len(b) }
func (b ByVersion) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByVersion) Less(i, j int) bool { return b[i].Version < b[j].Version }
