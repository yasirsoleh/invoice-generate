package repositories

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func OpenDatabase(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	err = RunMigrations(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}
