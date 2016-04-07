package darwin

import (
	"database/sql"
	"io"
)

type Migration struct {
	Version     float64
	Description string
	Script      io.Reader
}

type Dialect interface {
	CreateTableSQL() string
	MigrateSQL() string
	LastVersionSQL() string
}

type Darwin struct{}

func (d Darwin) Validate() bool {
	return false
}

func (d Darwin) Migrate() error {
	return nil
}

var DefaultDarwin = Darwin{}

func New(db *sql.DB, d Dialect, migrations []Migration) Darwin {
	return Darwin{}
}

func NewForMySQL(db *sql.DB, migrations []Migration) Darwin {
	return Darwin{}
}

func Validate(db *sql.DB, dialect Dialect, migrations []Migration) bool {
	return false
}

func Migrate(db *sql.DB, dialect Dialect, migrations []Migration) error {
	return nil
}

type MySQLDialect struct{}

func (m MySQLDialect) CreateTableSQL() string {
	return "CREATE TABLE schema_migrations"
}

func (m MySQLDialect) MigrateSQL() string {
	return "INSERT INTO schema_migrations"
}

func (m MySQLDialect) LastVersionSQL() string {
	return "SELECT version FROM schema_migrations ORDER BY version DESC"
}

type ByVersion []Migration

func (b ByVersion) Len() int           { return len(b) }
func (b ByVersion) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByVersion) Less(i, j int) bool { return b[i].Version < b[j].Version }
