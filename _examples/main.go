package main

import (
	"database/sql"
	"log"

	"github.com/GuiaBolso/darwin"
	_ "github.com/go-sql-driver/mysql"
)

var (
	migrations = []darwin.Migration{
		{
			Version:     1,
			Description: "Creating table posts",
			Script: `CREATE TABLE posts (
						id INT 		auto_increment, 
						title 		VARCHAR(255),
						PRIMARY KEY (id)
					 ) ENGINE=InnoDB CHARACTER SET=utf8;`,
		},
		{
			Version:     2,
			Description: "Adding column body",
			Script:      "ALTER TABLE posts ADD body TEXT AFTER title;",
		},
	}
)

func main() {
	database, err := sql.Open("mysql", "root:@/darwin")

	if err != nil {
		log.Fatal(err)
	}

	driver := darwin.NewGenericDriver(database, darwin.MySQLDialect{})

	d := darwin.New(driver, migrations, nil)
	err = d.Migrate()

	if err != nil {
		log.Println(err)
	}
}
