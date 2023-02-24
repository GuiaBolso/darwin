package generic_test

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ardanlabs/darwin/v2"
	"github.com/ardanlabs/darwin/v2/dialects/mysql"
	"github.com/ardanlabs/darwin/v2/drivers/generic"
)

func Test_GenericDriver_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New().error != nil, wants nil")
	}
	defer db.Close()

	var dialect mysql.Dialect

	mock.ExpectBegin()
	mock.ExpectExec(escapeQuery(dialect.CreateTableSQL())).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	d, err := generic.New(db, dialect)
	if err != nil {
		t.Fatalf("unable to construct driver: %s", err)
	}

	if err := d.Create(); err != nil {
		t.Fatalf("unable to create record: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expections: %s", err)
	}
}

func Test_GenericDriver_Insert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New().error != nil, wants nil")
	}
	defer db.Close()

	var dialect mysql.Dialect
	d, err := generic.New(db, dialect)
	if err != nil {
		t.Fatalf("unable to construct driver: %s", err)
	}

	record := darwin.MigrationRecord{
		Version:       1.0,
		Description:   "Description",
		Checksum:      "7ebca1c6f05333a728a8db4629e8d543",
		AppliedAt:     time.Now(),
		ExecutionTime: time.Millisecond * 1,
	}

	mock.ExpectBegin()
	mock.ExpectExec(escapeQuery(dialect.InsertSQL())).
		WithArgs(
			record.Version,
			record.Description,
			record.Checksum,
			record.AppliedAt.Unix(),
			record.ExecutionTime,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	if err := d.Insert(record); err != nil {
		t.Fatalf("unable to insert record: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expections: %s", err)
	}
}

func Test_GenericDriver_All_success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New().error != nil, wants nil")
	}
	defer db.Close()

	var dialect mysql.Dialect
	d, err := generic.New(db, dialect)
	if err != nil {
		t.Fatalf("unable to construct driver: %s", err)
	}

	rows := sqlmock.NewRows([]string{
		"version", "description", "checksum", "applied_at", "execution_time", "success",
	}).AddRow(
		1, "Description", "7ebca1c6f05333a728a8db4629e8d543",
		time.Now().Unix(),
		time.Millisecond*1, true,
	)

	mock.ExpectQuery(escapeQuery(dialect.AllSQL())).WillReturnRows(rows)

	migrations, err := d.All()
	if err != nil {
		t.Fatalf("unable to query record: %s", err)
	}

	if len(migrations) != 1 {
		t.Fatalf("len(migrations) == %d, wants 1", len(migrations))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expections: %s", err)
	}
}

func Test_GenericDriver_All_error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New().error != nil, wants nil")
	}
	defer db.Close()

	var dialect mysql.Dialect
	d, err := generic.New(db, dialect)
	if err != nil {
		t.Fatalf("unable to construct driver: %s", err)
	}

	mock.ExpectQuery(escapeQuery(dialect.AllSQL())).
		WillReturnError(errors.New("Generic error"))

	migrations, err := d.All()
	if err != nil {
		t.Fatalf("unable to query record: %s", err)
	}

	if len(migrations) != 0 {
		t.Fatalf("len(migrations) == %d, wants 0", len(migrations))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expections: %s", err)
	}
}

func Test_GenericDriver_Exec(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New().error != nil, wants nil")
	}
	defer db.Close()

	var dialect mysql.Dialect
	d, err := generic.New(db, dialect)
	if err != nil {
		t.Fatalf("unable to construct driver: %s", err)
	}

	stmt := "CREATE TABLE HELLO (id INT);"

	mock.ExpectBegin()
	mock.ExpectExec(escapeQuery(stmt)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if _, err := d.Exec(stmt); err != nil {
		t.Fatalf("unable to execute statement: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expections: %s", err)
	}
}

func Test_GenericDriver_Exec_error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("sqlmock.New().error != nil, wants nil")
	}
	defer db.Close()

	var dialect mysql.Dialect
	d, err := generic.New(db, dialect)
	if err != nil {
		t.Fatalf("unable to construct driver: %s", err)
	}

	stmt := "CREATE TABLE HELLO (id INT);"

	mock.ExpectBegin()
	mock.ExpectExec(escapeQuery(stmt)).
		WillReturnError(errors.New("Generic Error"))
	mock.ExpectRollback()

	d.Exec(stmt)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expections: %s", err)
	}
}

func Test_byMigrationRecordVersion(t *testing.T) {
	unordered := []darwin.MigrationRecord{
		{
			Version:       1.1,
			Description:   "Description",
			Checksum:      "7ebca1c6f05333a728a8db4629e8d543",
			AppliedAt:     time.Now(),
			ExecutionTime: time.Millisecond * 1,
		},
		{
			Version:       1.0,
			Description:   "Description",
			Checksum:      "7ebca1c6f05333a728a8db4629e8d543",
			AppliedAt:     time.Now(),
			ExecutionTime: time.Millisecond * 1,
		},
	}

	sort.Sort(byMigrationRecordVersion(unordered))

	if unordered[0].Version != 1.0 {
		t.Fatalf("Must order by version number")
	}
}

func Test_transaction_panic_sql_nil(t *testing.T) {
	f := func(tx *sql.Tx) error {
		return nil
	}

	err := transaction(nil, f)
	if err == nil {
		t.Fatalf("should not be able to execute a transaction with a db connection")
	}
}

func Test_transaction_error_begin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New().error != nil, wants nil")
	}
	defer db.Close()

	mock.ExpectBegin().WillReturnError(errors.New("Generic Error"))

	transaction(db, func(tx *sql.Tx) error {
		return nil
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expections: %s", err)
	}
}

func Test_transaction_panic_with_error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New().error != nil, wants nil")
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	err = transaction(db, func(tx *sql.Tx) error {
		panic(errors.New("Generic Error"))
	})

	if err == nil {
		t.Fatalf("Should handle the panic inside the transaction")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expections: %s", err)
	}
}

func Test_transaction_panic_with_message(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New().error != nil, wants nil")
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	err = transaction(db, func(tx *sql.Tx) error {
		panic("Generic Error")
	})

	if err == nil {
		t.Fatalf("Should handle the panic inside the transaction")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expections: %s", err)
	}
}

func escapeQuery(s string) string {
	re := regexp.MustCompile(`\\s+`)

	s1 := regexp.QuoteMeta(s)
	s1 = strings.TrimSpace(re.ReplaceAllString(s1, " "))
	return s1
}

// =============================================================================

// transaction is a utility function to execute the SQL inside a transaction.
// see: http://stackoverflow.com/a/23502629
func transaction(db *sql.DB, f func(*sql.Tx) error) (err error) {
	if db == nil {
		return errors.New("darwin: sql.DB is nil")
	}

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

// =============================================================================

type byMigrationRecordVersion []darwin.MigrationRecord

func (b byMigrationRecordVersion) Len() int           { return len(b) }
func (b byMigrationRecordVersion) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byMigrationRecordVersion) Less(i, j int) bool { return b[i].Version < b[j].Version }
