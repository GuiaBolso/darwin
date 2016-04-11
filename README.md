[![Build Status](https://travis-ci.org/dimiro1/darwin.svg?branch=master)](https://travis-ci.org/dimiro1/darwin)
[![Go Report Card](https://goreportcard.com/badge/github.com/dimiro1/darwin)](https://goreportcard.com/report/github.com/dimiro1/darwin)
[![GoDoc](https://godoc.org/github.com/dimiro1/darwin?status.svg)](https://godoc.org/github.com/dimiro1/darwin)

# Darwin

Database schema evolution library for Go

# Example

```go
package main

import (
	"database/sql"
	"log"

	"github.com/dimiro1/darwin"
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
```

# LICENSE

The MIT License (MIT)

Copyright (c) 2016 Claudemiro

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.