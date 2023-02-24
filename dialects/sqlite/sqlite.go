// Package sqlite provides support to work with a sqlite database.
package sqlite

// Dialect a Dialect configured for Sqlite3.
type Dialect struct{}

// CreateTableSQL returns the SQL to create the schema table.
func (Dialect) CreateTableSQL() string {
	return `CREATE TABLE IF NOT EXISTS darwin_migrations
                (
                    id             INTEGER  PRIMARY KEY,
                    version        FLOAT    NOT NULL,
                    description    TEXT     NOT NULL,
                    checksum       TEXT     NOT NULL,
                    applied_at     DATETIME NOT NULL,
                    execution_time FLOAT    NOT NULL,
                    UNIQUE         (version)
                );`
}

// InsertSQL returns the SQL to insert a new migration in the schema table.
func (Dialect) InsertSQL() string {
	return `INSERT INTO darwin_migrations
                (
                    version,
                    description,
                    checksum,
                    applied_at,
                    execution_time
                )
            VALUES (?, ?, ?, ?, ?);`
}

// UpdateChecksumSQL returns the SQL update a checksum for a version.
func (Dialect) UpdateChecksumSQL() string {
	return `UPDATE darwin_migrations
			SET
				checksum = ?
			WHERE
				version = ?;`
}

// AllSQL returns a SQL to get all entries in the table.
func (Dialect) AllSQL() string {
	return `SELECT 
                version,
                description,
                checksum,
                applied_at,
                execution_time
            FROM 
                darwin_migrations
            ORDER BY version ASC;`
}
