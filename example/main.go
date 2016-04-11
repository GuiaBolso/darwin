package main

import (
	"database/sql"
	"log"

	"github.com/dimiro1/darwin"
	_ "github.com/go-sql-driver/mysql"
)

// https://flywaydb.org/documentation/maven/validate
var (
	migrations = []darwin.Migration{
		{
			Version:     1.0,
			Description: "Creating table people",
			Script:      "CREATE TABLE people (id INT);",
		},
		{
			Version:     1.1,
			Description: "Creating table animals",
			Script:      "CREATE TABLE animals (id INT);",
		},
		{
			Version:     2,
			Description: "Creating table hello",
			Script:      "CREATE TABLE hello (id INT);",
		},
		{
			Version:     3,
			Description: "Creating table world",
			Script:      "CREATE TABLE world (id INT);",
		},
		// {
		// 	Version:     4,
		// 	Description: "Creating table ok",
		// 	Script:      "CREATE TABLE ok (id INT);",
		// },
	}
)

func main() {
	database, err := sql.Open("mysql", "root:@/darwin")

	if err != nil {
		log.Println(err)
		return
	}

	driver := darwin.NewGenericDriver(database, darwin.MySQLDialect{})

	d := darwin.New(driver, migrations, nil)
	err = d.Migrate()

	if err != nil {
		log.Println(err)
	}
}
