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

func (m Migration) Less(other Migration) bool {
	return m.Version < other.Version
}

type Dialect interface {
	CreateTableSQL() string
	MigrateSQL() string
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

type Migrations []Migration

func (ml Migrations) Len() int           { return len(ml) }
func (ml Migrations) Less(i, j int) bool { return ml[i].Less(ml[j]) }
func (ml Migrations) Swap(i, j int)      { ml[i], ml[j] = ml[j], ml[i] }
