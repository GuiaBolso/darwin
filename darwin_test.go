package darwin

import (
	"errors"
	"fmt"
	"sort"
	"testing"
	"time"
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

func Test_Status_String(t *testing.T) {
	expectations := []struct {
		status   Status
		expected string
	}{
		{
			Ignored, "IGNORED",
		},
		{
			Applied, "APPLIED",
		},
		{
			Pending, "PENDING",
		},
		{
			Error, "ERROR",
		},
		{
			Status(-1), "INVALID",
		},
	}

	for _, expectation := range expectations {
		if expectation.expected != expectation.status.String() {
			t.Errorf("Expected %s, got %s", expectation.expected, expectation.status.String())
			t.FailNow()
		}
	}
}

func Test_Info(t *testing.T) {
	baseTime, _ := time.Parse(time.RFC3339, "2002-10-02T15:00:00Z")

	records := []MigrationRecord{
		{
			Version:     1.0,
			Description: "1.0",
			AppliedAt:   baseTime,
		},
		{
			Version:     2.0,
			Description: "2.0",
			AppliedAt:   baseTime.Add(2 * time.Second),
		},
	}

	migrations := []Migration{
		{
			Version:     1.0,
			Description: "Must Be APPLIED",
			Script:      "does not matter!",
		},
		{
			Version:     1.1,
			Description: "Must Be IGNORED",
			Script:      "does not matter!",
		},
		{
			Version:     2.0,
			Description: "Must Be APPLIED",
			Script:      "does not matter!",
		},
		{
			Version:     3.0,
			Description: "Must Be PENDING",
			Script:      "does not matter!",
		},
	}

	d := New(&dummyDriver{records: records}, migrations)
	d.Migrate()
	infos, err := d.Info()

	if err != nil {
		t.Error("Must not return error")
		t.FailNow()
	}

	expectations := []Status{Applied, Ignored, Applied, Pending}

	for i, info := range infos {
		if expectations[i] != info.Status {
			t.Errorf("Expected %s, got %s", expectations[i], info.Status)
			t.FailNow()
		}
	}
}

func Test_Info_with_error(t *testing.T) {
	d := New(&dummyDriver{AllError: true}, []Migration{})

	if _, err := d.Info(); err == nil {
		t.Error("Must emit error")
	}
}

func Test_DuplicateMigrationVersionError_Error(t *testing.T) {
	err := DuplicateMigrationVersionError{Version: 1}

	if err.Error() != fmt.Sprintf("Multiple migrations have the version number %f.", 1.0) {
		t.Error("Must inform the version of the duplicated migration")
	}
}

func Test_IllegalMigrationVersionError_Error(t *testing.T) {
	err := IllegalMigrationVersionError{Version: 1}

	if err.Error() != fmt.Sprintf("Illegal migration version number %f.", 1.0) {
		t.Error("Must inform the version of the invalid migration")
	}
}

func Test_RemovedMigrationError_Error(t *testing.T) {
	err := RemovedMigrationError{Version: 1}

	if err.Error() != fmt.Sprintf("Migration %f was removed", 1.0) {
		t.Error("Must inform when a migration is removed from the list")
	}
}

func Test_InvalidChecksumError_Error(t *testing.T) {
	err := InvalidChecksumError{Version: 1}

	if err.Error() != fmt.Sprintf("Invalid cheksum for migration %f", 1.0) {
		t.Error("Must inform when a migration have an invalid checksum")
	}
}

func Test_Validate_invalid_version(t *testing.T) {
	migrations := []Migration{
		{
			Version:     -1,
			Description: "Hello World",
			Script:      "does not matter!",
		},
	}

	d := New(&dummyDriver{}, migrations)

	err := d.Validate()
	if err.(*IllegalMigrationVersionError).Version != -1 {
		t.Errorf("Must not accept migrations with invalid version numbers")
	}
}

func Test_Validate_duplicated_version(t *testing.T) {
	migrations := []Migration{
		{
			Version:     1,
			Description: "Hello World",
			Script:      "does not matter!",
		},
		{
			Version:     1,
			Description: "Hello World",
			Script:      "does not matter!",
		},
	}

	d := New(&dummyDriver{}, migrations)

	err := d.Validate()
	if err.(*DuplicateMigrationVersionError).Version != 1 {
		t.Errorf("Must not accept migrations with duplicated version numbers")
	}
}

func Test_Validate_removed_migration(t *testing.T) {
	records := []MigrationRecord{
		{
			Version: 1.0,
		},
		{
			Version: 1.1,
		},
	}

	migrations := []Migration{
		{
			Version:     1.1,
			Description: "Hello World",
			Script:      "does not matter!",
		},
	}

	d := New(&dummyDriver{records: records}, migrations)

	err := d.Validate()
	if err.(*RemovedMigrationError).Version != 1 {
		t.Errorf("Must not validate when some migration was removed from the migration list")
	}
}

func Test_Validate_invalid_checksum(t *testing.T) {
	records := []MigrationRecord{
		{
			Version:  1.0,
			Checksum: "3310d0ff858faac79e854454c9e403db",
		},
	}

	migrations := []Migration{
		{
			Version:     1.0,
			Description: "Hello World",
			Script:      "does not matter!",
		},
	}

	d := New(&dummyDriver{records: records}, migrations)

	err := d.Validate()
	if err.(*InvalidChecksumError).Version != 1 {
		t.Errorf("Must not validate when some migration differ from the migration applied in the database")
	}
}

func Test_Migrate_migrate_all(t *testing.T) {
	migrations := []Migration{
		{
			Version:     1,
			Description: "First Migration",
			Script:      "does not matter!",
		},
		{
			Version:     2,
			Description: "Second Migration",
			Script:      "does not matter!",
		},
	}

	d := New(&dummyDriver{records: []MigrationRecord{}}, migrations)
	d.Migrate()

	all, _ := d.driver.All()
	if len(all) != 2 {
		t.Errorf("Must not apply all migrations")
	}
}

func Test_Migrate_migrate_partial(t *testing.T) {
	applied := []MigrationRecord{
		{
			Version:  1,
			Checksum: "3310d0ff858faac79e854454c9e403da",
		},
	}

	migrations := []Migration{
		{
			Version:     1,
			Description: "First Migration",
			Script:      "does not matter!",
		},
		{
			Version:     2,
			Description: "Second Migration",
			Script:      "does not matter!",
		},
		{
			Version:     3,
			Description: "Third Migration",
			Script:      "does not matter!",
		},
	}

	driver := &dummyDriver{records: applied}

	all, _ := driver.All()

	if len(all) != 1 {
		t.Errorf("Should have 1 migration already applied")
	}

	// Running with struct
	d := New(driver, migrations)
	d.Migrate()

	all, _ = driver.All()

	if len(all) != 3 {
		t.Errorf("Must not apply all migrations")
	}
}

func Test_Migrate_migrate_error(t *testing.T) {
	d := New(&dummyDriver{CreateError: true}, []Migration{})

	err := d.Migrate()
	if err == nil {
		t.Error("Must emit error")
	}
}

func Test_Migrate_with_error_in_Validate(t *testing.T) {
	d := New(&dummyDriver{AllError: true}, []Migration{})

	err := d.Migrate()
	if err == nil {
		t.Error("Must emit error")
	}
}

func Test_Migrate_with_error_in_driver_insert(t *testing.T) {
	migrations := []Migration{
		{
			Version:     1,
			Description: "First Migration",
			Script:      "does not matter!",
		},
	}

	d := New(&dummyDriver{InsertError: true}, migrations)

	err := d.Migrate()
	if err == nil {
		t.Error("Must emit error")
	}
}

func Test_Migrate_with_error_in_driver_exec(t *testing.T) {
	migrations := []Migration{
		{
			Version:     1,
			Description: "First Migration",
			Script:      "does not matter!",
		},
	}

	d := New(&dummyDriver{ExecError: true}, migrations)
	d.Migrate()

	all, _ := d.driver.All()
	if len(all) != 0 {
		t.Errorf("Must not apply all migrations")
	}
}

func Test_planMigration_error_driver(t *testing.T) {
	d := New(&dummyDriver{AllError: true}, []Migration{})

	_, err := d.planMigration()
	if err == nil {
		t.Error("Must emit error")
	}
}

func Test_byMigrationVersion(t *testing.T) {
	unordered := []Migration{
		{
			Version:     3,
			Description: "Hello World",
			Script:      "does not matter!",
		},
		{
			Version:     1,
			Description: "Hello World",
			Script:      "does not matter!",
		},
	}

	sort.Sort(byMigrationVersion(unordered))

	if unordered[0].Version != 1.0 {
		t.Errorf("Must order by version number")
	}
}

func TestParse(t *testing.T) {
	t.Log("Given the need to parse a sql migration file.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling the embedded schema.", testID)
		{
			migs := ParseMigrations(schemaDoc)
			for i, mig := range migs {
				if mig.Checksum() != checksumDoc[i] {
					t.Log("got:", mig.Checksum())
					t.Log("exp:", checksumDoc[i])
					t.Errorf("\t%s\tTest %d:\tShould have correct checksum for version %f.", failed, testID, mig.Version)
				}
			}
		}
	}
}

// =============================================================================

type dummyDriver struct {
	CreateError bool
	InsertError bool
	UpdateError bool
	AllError    bool
	ExecError   bool
	records     []MigrationRecord
}

func (d *dummyDriver) Create() error {
	if d.CreateError {
		return errors.New("Error")
	}
	return nil
}

func (d *dummyDriver) Insert(m MigrationRecord) error {
	if d.InsertError {
		return errors.New("Error")
	}

	d.records = append(d.records, m)
	return nil
}

func (d *dummyDriver) UpdateChecksum(checksum string, version float64) error {
	if d.UpdateError {
		return errors.New("Error")
	}

	for i, record := range d.records {
		if record.Version == version {
			record.Checksum = checksum
			d.records[i] = record
		}
	}

	return nil
}

func (d *dummyDriver) All() ([]MigrationRecord, error) {
	if d.AllError {
		return []MigrationRecord{}, errors.New("Error")
	}

	return d.records, nil
}

func (d *dummyDriver) Exec(string) (time.Duration, error) {
	if d.ExecError {
		return time.Millisecond * 1, errors.New("Error")
	}

	return time.Millisecond * 1, nil
}

// =============================================================================

var checksumDoc = []string{
	"f06593a9b87baa8fcad94582aad566e7",
	"6b37bb9170d4d44e609ded3776eb3a32",
	"7694752761db238c2ea9290267430ad6",
	"5f840a746f43e417ce86eb9c46bb252b",
}

var schemaDoc = `-- Version: 1.1
-- Description: Create table users
CREATE TABLE users (
	user_id       UUID,
	name          TEXT,
	email         TEXT UNIQUE,
	roles         TEXT[],
	password_hash TEXT,
	date_created  TIMESTAMP,
	date_updated  TIMESTAMP,

	PRIMARY KEY (user_id)
);

-- Version: 1.2
-- Description: Create table products
CREATE TABLE products (
	product_id   UUID,
	name         TEXT,
	cost         INT,
	quantity     INT,
	date_created TIMESTAMP,
	date_updated TIMESTAMP,

	PRIMARY KEY (product_id)
);

-- Version: 1.3
-- Description: Create table sales
CREATE TABLE sales (
	sale_id      UUID,
	product_id   UUID,
	quantity     INT,
	paid         INT,
	date_created TIMESTAMP,

	PRIMARY KEY (sale_id),
	FOREIGN KEY (product_id) REFERENCES products(product_id) ON DELETE CASCADE
);

-- Version: 2.1
-- Description: Alter table products with user column"
ALTER TABLE products
	ADD COLUMN user_id UUID DEFAULT '00000000-0000-0000-0000-000000000000'
`
