package darwin

import (
	"database/sql"
	"fmt"
	"time"
)

// MigrationRecord is the entry in schema table
type MigrationRecord struct {
	Version       float64
	Description   string
	Checksum      string
	AppliedAt     time.Time
	ExecutionTime time.Duration
	Success       bool
}

// Driver a database driver abstraction
type Driver interface {
	Create() error
	Insert(e MigrationRecord) error
	All() ([]MigrationRecord, error)
	Exec(string) (time.Duration, error)
}

// GenericDriver is the default Driver, it can be configured to any database.
type GenericDriver struct {
	db      *sql.DB
	dialect Dialect
}

// NewGenericDriver creates a new GenericDriver configured with db and dialect
func NewGenericDriver(db *sql.DB, dialect Dialect) *GenericDriver {
	return &GenericDriver{db: db, dialect: dialect}
}

// Create create the table darwin_migrations if necessary
func (m *GenericDriver) Create() error {
	err := transaction(m.db, func(tx *sql.Tx) error {
		_, err := tx.Exec(m.dialect.CreateTableSQL())
		return err
	})

	return err
}

// Insert insert a migration entry into database
func (m *GenericDriver) Insert(e MigrationRecord) error {

	err := transaction(m.db, func(tx *sql.Tx) error {
		_, err := tx.Exec(m.dialect.InsertSQL(),
			e.Version,
			e.Description,
			e.Checksum,
			e.AppliedAt.Unix(),
			e.ExecutionTime,
			e.Success,
		)
		return err
	})

	return err
}

// All returns all migrations applied
func (m *GenericDriver) All() ([]MigrationRecord, error) {
	entries := []MigrationRecord{}

	rows, err := m.db.Query(m.dialect.AllSQL())

	if err != nil {
		return []MigrationRecord{}, err
	}

	for rows.Next() {
		var (
			version       float64
			description   string
			checksum      string
			appliedAt     int64
			executionTime float64
			success       bool
		)

		rows.Scan(
			&version,
			&description,
			&checksum,
			&appliedAt,
			&executionTime,
			&success,
		)

		entry := MigrationRecord{
			Version:       version,
			Description:   description,
			Checksum:      checksum,
			AppliedAt:     time.Unix(appliedAt, 0),
			ExecutionTime: time.Duration(executionTime),
			Success:       success,
		}

		entries = append(entries, entry)
	}

	rows.Close()

	return entries, nil
}

// Exec execute sql scripts into database
func (m *GenericDriver) Exec(script string) (time.Duration, error) {
	start := time.Now()

	err := transaction(m.db, func(tx *sql.Tx) error {
		_, err := tx.Exec(script)
		return err
	})

	return time.Since(start), err
}

// transaction is a utility function to execute the SQL inside a transaction
// see: http://stackoverflow.com/a/23502629
func transaction(db *sql.DB, f func(*sql.Tx) error) (err error) {
	tx, err := db.Begin()

	if err != nil {
		return
	}

	defer func() {
		if p := recover(); p != nil {
			switch p := p.(type) {
			case error:
				err = p
			default:
				err = fmt.Errorf("%s", p)
			}
		}
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	return f(tx)
}

type byMigrationRecordVersion []MigrationRecord

func (b byMigrationRecordVersion) Len() int           { return len(b) }
func (b byMigrationRecordVersion) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byMigrationRecordVersion) Less(i, j int) bool { return b[i].Version < b[j].Version }
