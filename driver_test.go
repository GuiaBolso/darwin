package darwin

import (
	"database/sql"
	"errors"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func assertPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}

func Test_NewGenericDriver_sql_nil(t *testing.T) {
	assertPanic(t, func() {
		NewGenericDriver(nil, MySQLDialect{})
	})
}

func Test_NewGenericDriver_driver_nil(t *testing.T) {
	db, _, _ := sqlmock.New()

	defer db.Close()

	assertPanic(t, func() {
		NewGenericDriver(db, nil)
	})
}

func Test_GenericDriver_Create(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Errorf("sqlmock.New().error != nil, wants nil")
	}

	defer db.Close()

	dialect := MySQLDialect{}

	mock.ExpectBegin()
	mock.ExpectExec(escapeQuery(dialect.CreateTableSQL())).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	d := NewGenericDriver(db, dialect)
	d.Create()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func Test_GenericDriver_Insert(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Errorf("sqlmock.New().error != nil, wants nil")
	}

	defer db.Close()

	record := MigrationRecord{
		Version:       1.0,
		Description:   "Description",
		Checksum:      "7ebca1c6f05333a728a8db4629e8d543",
		AppliedAt:     time.Now(),
		ExecutionTime: time.Millisecond * 1,
	}

	dialect := MySQLDialect{}

	d := NewGenericDriver(db, dialect)

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

	d.Insert(record)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func Test_GenericDriver_All_success(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Errorf("sqlmock.New().error != nil, wants nil")
	}

	defer db.Close()

	dialect := MySQLDialect{}

	d := NewGenericDriver(db, dialect)

	rows := sqlmock.NewRows([]string{
		"version", "description", "checksum", "applied_at", "execution_time", "success",
	}).AddRow(
		1, "Description", "7ebca1c6f05333a728a8db4629e8d543",
		time.Now().Unix(),
		time.Millisecond*1, true,
	)

	mock.ExpectQuery(escapeQuery(dialect.AllSQL())).
		WillReturnRows(rows)

	migrations, _ := d.All()

	if len(migrations) != 1 {
		t.Errorf("len(migrations) == %d, wants 1", len(migrations))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func Test_GenericDriver_All_error(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Errorf("sqlmock.New().error != nil, wants nil")
	}

	defer db.Close()

	dialect := MySQLDialect{}

	d := NewGenericDriver(db, dialect)

	mock.ExpectQuery(escapeQuery(dialect.AllSQL())).
		WillReturnError(errors.New("Generic error"))

	migrations, _ := d.All()

	if len(migrations) != 0 {
		t.Errorf("len(migrations) == %d, wants 0", len(migrations))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func Test_GenericDriver_Exec(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Errorf("sqlmock.New().error != nil, wants nil")
	}

	defer db.Close()

	stmt := "CREATE TABLE HELLO (id INT);"
	dialect := MySQLDialect{}

	d := NewGenericDriver(db, dialect)

	mock.ExpectBegin()
	mock.ExpectExec(escapeQuery(stmt)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	d.Exec(stmt)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func Test_GenericDriver_Exec_error(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Errorf("sqlmock.New().error != nil, wants nil")
	}

	defer db.Close()

	stmt := "CREATE TABLE HELLO (id INT);"
	dialect := MySQLDialect{}

	d := NewGenericDriver(db, dialect)

	mock.ExpectBegin()
	mock.ExpectExec(escapeQuery(stmt)).
		WillReturnError(errors.New("Generic Error"))
	mock.ExpectRollback()

	d.Exec(stmt)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func Test_byMigrationRecordVersion(t *testing.T) {
	unordered := []MigrationRecord{
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
		t.Errorf("Must order by version number")
	}
}

func Test_transaction_panic_sql_nil(t *testing.T) {
	assertPanic(t, func() {
		transaction(nil, func(tx *sql.Tx) error {
			return nil
		})
	})
}

func Test_transaction_error_begin(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Errorf("sqlmock.New().error != nil, wants nil")
	}

	defer db.Close()

	mock.ExpectBegin().WillReturnError(errors.New("Generic Error"))

	transaction(db, func(tx *sql.Tx) error {
		return nil
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func Test_transaction_panic_with_error(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Errorf("sqlmock.New().error != nil, wants nil")
	}

	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	err = transaction(db, func(tx *sql.Tx) error {
		panic(errors.New("Generic Error"))
	})

	if err == nil {
		t.Errorf("Should handle the panic inside the transaction")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func Test_transaction_panic_with_message(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Errorf("sqlmock.New().error != nil, wants nil")
	}

	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	err = transaction(db, func(tx *sql.Tx) error {
		panic("Generic Error")
	})

	if err == nil {
		t.Errorf("Should handle the panic inside the transaction")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func escapeQuery(s string) string {
	re := regexp.MustCompile("\\s+")

	s1 := regexp.QuoteMeta(s)
	s1 = strings.TrimSpace(re.ReplaceAllString(s1, " "))
	return s1
}
