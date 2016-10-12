package darwin

import (
	"errors"
	"fmt"
	"sort"
	"testing"
	"time"
)

type dummyDriver struct {
	CreateError bool
	InsertError bool
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

	d := New(&dummyDriver{records: records}, migrations, nil)
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
	driver := &dummyDriver{AllError: true}
	migrations := []Migration{}

	_, err := Info(driver, migrations)

	if err == nil {
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

	err := Validate(&dummyDriver{}, migrations)

	if err.(IllegalMigrationVersionError).Version != -1 {
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

	err := Validate(&dummyDriver{}, migrations)

	if err.(DuplicateMigrationVersionError).Version != 1 {
		t.Errorf("Must not accept migrations with duplicated version numbers")
	}
}

func Test_Validate_removed_migration(t *testing.T) {
	// Other fields are not necessary for testing...
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

	// Running with struct
	d := New(&dummyDriver{records: records}, migrations, nil)
	err := d.Validate()

	if err.(RemovedMigrationError).Version != 1 {
		t.Errorf("Must not validate when some migration was removed from the migration list")
	}
}

func Test_Validate_invalid_checksum(t *testing.T) {
	// Other fields are not necessary for testing...
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

	err := Validate(&dummyDriver{records: records}, migrations)

	if err.(InvalidChecksumError).Version != 1 {
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

	driver := &dummyDriver{records: []MigrationRecord{}}

	infoChan := make(chan MigrationInfo, 2)

	Migrate(driver, migrations, infoChan)

	all, _ := driver.All()

	if len(all) != 2 {
		t.Errorf("Must not apply all migrations")
	}

	info := <-infoChan

	if info.Migration.Version != 1 {
		t.Errorf("Must send a message for each migration applied")
	}

	info = <-infoChan

	if info.Migration.Version != 2 {
		t.Errorf("Must send a message for each migration applied")
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
	d := New(driver, migrations, nil)
	d.Migrate()

	all, _ = driver.All()

	if len(all) != 3 {
		t.Errorf("Must not apply all migrations")
	}
}

func Test_Migrate_migrate_error(t *testing.T) {
	driver := &dummyDriver{CreateError: true}
	migrations := []Migration{}

	err := Migrate(driver, migrations, nil)

	if err == nil {
		t.Error("Must emit error")
	}
}

func Test_Migrate_with_error_in_Validate(t *testing.T) {
	driver := &dummyDriver{AllError: true}
	migrations := []Migration{}

	err := Migrate(driver, migrations, nil)

	if err == nil {
		t.Error("Must emit error")
	}
}

func Test_Migrate_with_error_in_driver_insert(t *testing.T) {
	driver := &dummyDriver{InsertError: true}
	migrations := []Migration{
		{
			Version:     1,
			Description: "First Migration",
			Script:      "does not matter!",
		},
	}

	err := Migrate(driver, migrations, nil)

	if err == nil {
		t.Error("Must emit error")
	}
}

func Test_Migrate_with_error_in_driver_exec(t *testing.T) {
	driver := &dummyDriver{ExecError: true}
	migrations := []Migration{
		{
			Version:     1,
			Description: "First Migration",
			Script:      "does not matter!",
		},
	}

	Migrate(driver, migrations, nil)

	all, _ := driver.All()

	if len(all) != 0 {
		t.Errorf("Must not apply all migrations")
	}
}

func Test_planMigration_error_driver(t *testing.T) {
	driver := &dummyDriver{AllError: true}
	migrations := []Migration{}

	_, err := planMigration(driver, migrations)

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
