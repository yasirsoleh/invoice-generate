package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yasirsoleh/invoice-generate/models"
)

func NextInvoiceNumber(tx *sql.Tx, date time.Time) (string, error) {
	dateStr := date.Format("2006-01-02")

	var last int
	err := tx.QueryRow(`
		SELECT last_number
		FROM invoice_counters
		WHERE invoice_date = ?
	`, dateStr).Scan(&last)

	switch err {
	case sql.ErrNoRows:
		last = 1
		_, err = tx.Exec(`
			INSERT INTO invoice_counters (invoice_date, last_number)
			VALUES (?, ?)
		`, dateStr, last)
	case nil:
		last++
		_, err = tx.Exec(`
			UPDATE invoice_counters
			SET last_number = ?
			WHERE invoice_date = ?
		`, last, dateStr)
	}

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("INV-%s-%03d", date.Format("20060102"), last), nil
}

func InsertInvoice(tx *sql.Tx, invoiceNumber string, invoiceData models.Invoice) (string, error) {
	dataJSON, err := json.Marshal(invoiceData)
	if err != nil {
		return "", err
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	_, err = tx.Exec(`
		INSERT INTO invoices (id, invoice_number, invoice_data)
		VALUES (?, ?, ?)
	`, id.String(), invoiceNumber, dataJSON)
	return id.String(), err
}

func GetInvoice(db *sql.DB, id string) (models.Invoice, error) {
	var dataJSON []byte
	err := db.QueryRow(`
		SELECT invoice_data
		FROM invoices
		WHERE id = ?
	`, id).Scan(&dataJSON)
	if err != nil {
		return models.Invoice{}, err
	}

	var invoice models.Invoice
	err = json.Unmarshal(dataJSON, &invoice)
	if err != nil {
		return models.Invoice{}, err
	}

	return invoice, nil
}
