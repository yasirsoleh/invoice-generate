package repositories

import "database/sql"

const migration = `
CREATE TABLE IF NOT EXISTS invoice_counters (
	invoice_date TEXT PRIMARY KEY, -- YYYY-MM-DD
	last_number INTEGER NOT NULL
);

-- store every invoice generated using json
CREATE TABLE IF NOT EXISTS invoices (
	id TEXT UNIQUE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	invoice_number TEXT PRIMARY KEY,
	invoice_data JSON NOT NULL
);
`

func RunMigrations(db *sql.DB) error {
	_, err := db.Exec(migration)
	return err
}
