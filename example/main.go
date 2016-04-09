package main

import (
	"database/sql"
	"log"
	"strings"

	"github.com/dimiro1/darwin"
	_ "github.com/go-sql-driver/mysql"
)

/*
 table darwin_schema_version

 id, version, description, checksums, execution_time, success
*/

// Validar Version
// Não pode ter duas migrations com o mesmo Version, erro duplicada
// Unknown Version, caso Version is nil
// Version is a float
// Verbose mode - Print in StdOut or Writer every uteraction
// ILegal migrations < 0
// Valid version numbers are 1, 2, 3, 3.1 etc
// migrations are sorted by this version number
// Build a Migration plan
// Execute the plan.
// https://flywaydb.org/documentation/maven/validate
var (
	migrations = []darwin.Migration{
		darwin.Migration{
			Version:     1.0,
			Description: "Creating table people",
			Script:      strings.NewReader("CREATE TABLE people (id INT);"),
		},
		darwin.Migration{
			Version:     1.1,
			Description: "Creating table animals",
			Script:      strings.NewReader("CREATE TABLE animals (id INT);"),
		},
	}
)

func main() {
	database, err := sql.Open("mysql", "root:@/darwin")

	if err != nil {
		log.Println(err)
		return
	}

	err = darwin.Migrate(database, darwin.MySQLDialect{}, migrations)

	if err != nil {
		log.Println(err)
		return
	}
}
