package main

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/dimiro1/darwin"
)

/*
 table darwin_schema_version

 id, version, description, checksums, execution_time, success
*/

// Validar Version
// Não pode ter duas migrations com o mesmo Version
// Version is a float
// Valid version numbers are 1, 2, 3, 3.1 etc
// migrations are sorted by this version number
// https://flywaydb.org/documentation/maven/validate
var (
	migrations = []darwin.Migration{
		darwin.Migration{
			Version:     1.0,
			Description: "Creating table people",
			Script:      strings.NewReader("CREATE TABLE people (id int)"),
		},
	}
)

func main() {
	database, err := sql.Open("mysql", "/")

	if err != nil {
		// Handle errors!
	}

	m := darwin.NewForMySQL(database, migrations)
	m.Validate()
	m.Migrate()

	ok := darwin.Validate(database, darwin.MySQLDialect{}, migrations)

	fmt.Println(ok)

	err = darwin.Migrate(database, darwin.MySQLDialect{}, migrations)

	if err != nil {
		// Handle errors!
	}
}
