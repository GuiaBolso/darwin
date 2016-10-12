package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/GuiaBolso/darwin"
	_ "github.com/mattn/go-sqlite3"
)

var (
	migrations = []darwin.Migration{
		{
			Version:     1,
			Description: "Creating table posts",
			Script: `CREATE TABLE posts (
						id INTEGER PRIMARY KEY, 
						title 		TEXT
					 );;`,
		},
		{
			Version:     2,
			Description: "Adding column body",
			Script:      "ALTER TABLE posts ADD body TEXT AFTER title;",
		},
	}
)

func main() {
	var info bool

	flag.BoolVar(&info, "info", false, "If you want get info from database")
	flag.Parse()

	database, err := sql.Open("sqlite3", "database.sqlite3")

	if err != nil {
		log.Fatal(err)
	}

	driver := darwin.NewGenericDriver(database, darwin.SqliteDialect{})

	d := darwin.New(driver, migrations, nil)

	if info {
		infos, _ := d.Info()
		for _, info := range infos {
			fmt.Printf("%.1f %s %s\n", info.Migration.Version, info.Status, info.Migration.Description)
		}
	} else {
		err = d.Migrate()

		if err != nil {
			log.Println(err)
		}
	}
}
