package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/yasirsoleh/invoice-generate/models"
	"github.com/yasirsoleh/invoice-generate/pages"
	"github.com/yasirsoleh/invoice-generate/repositories"
)

func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, `\r\n`, "\n")
	s = strings.ReplaceAll(s, `\n`, "\n")
	return s
}

func getEnvOrFlag(envKey, flagValue string) string {
	if envValue, exists := os.LookupEnv(envKey); exists {
		return envValue
	}
	return flagValue
}

func getEnvOrFlagFloat(envKey string, flagValue float64) float64 {
	if envValue, exists := os.LookupEnv(envKey); exists {
		var f float64
		_, err := fmt.Sscanf(envValue, "%f", &f)
		if err == nil {
			return f
		}
	}
	return flagValue
}

func main() {
	// load cli args here for issuer details, recipient details, items, etc.
	serveAddress := flag.String("serve-address", ":8080", "Address to serve the web application on")
	dbPath := flag.String("db-path", "invoices.db", "Path to the SQLite database file")
	issuerName := flag.String("issuer-name", "", "Name of the invoice issuer")
	issuerNRIC := flag.String("issuer-nric", "", "NRIC of the invoice issuer")
	issuerBRN := flag.String("issuer-brn", "", "BRN of the invoice issuer")
	issuerTaxNumber := flag.String("issuer-tax", "", "Tax number of the invoice issuer")
	issuerAddress := flag.String("issuer-address", "", "Address of the invoice issuer")
	issuerEmail := flag.String("issuer-email", "", "Email of the invoice issuer")
	issuerPhone := flag.String("issuer-phone", "", "Phone number of the invoice issuer")
	paymentMethod := flag.String("payment-method", "Bank Transfer", "Payment method")
	paymentBankName := flag.String("payment-bank", "", "Bank name for payment")
	paymentAccountName := flag.String("payment-account-name", "", "Account name for payment")
	paymentAccountNumber := flag.String("payment-account-number", "", "Account number for payment")
	currencyCode := flag.String("currency_code", "RM", "Currency code for the invoice")
	taxRate := flag.Float64("tax-rate", 0.06, "Tax rate as a decimal (e.g., 0.06 for 6%)")

	flag.Parse()

	issuerAddressNormalized := normalizeNewlines(getEnvOrFlag("ISSUER_ADDRESS", *issuerAddress))

	// open db
	db, err := repositories.OpenDatabase(getEnvOrFlag("DB_PATH", *dbPath))
	if err != nil {
		panic(err)
	}

	// serve create invoice form
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("invoice form create request from: %s\n", r.RemoteAddr)
		pages.CreateInvoicePage().Render(r.Context(), w)
	})

	// handle create invoice form submission
	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("invoice form submit request from: %s\n", r.RemoteAddr)
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		date, receipient, items, err := pages.ParseInvoice(r)
		if err != nil {
			http.Error(w, "Failed to parse invoice: "+err.Error(), http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Failed to start transaction: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// generate invoice number
		invNumber, err := repositories.NextInvoiceNumber(tx, date)
		if err != nil {
			http.Error(w, "Failed to generate invoice number: "+err.Error(), http.StatusInternalServerError)
			return
		}

		inv := models.Invoice{
			Number: invNumber,
			Date:   date.Format("2006-01-02"),
			Issuer: models.Party{
				Name:      getEnvOrFlag("ISSUER_NAME", *issuerName),
				NRIC:      getEnvOrFlag("ISSUER_NRIC", *issuerNRIC),
				BRN:       getEnvOrFlag("ISSUER_BRN", *issuerBRN),
				TaxNumber: getEnvOrFlag("ISSUER_TAX", *issuerTaxNumber),
				Address:   issuerAddressNormalized,
				Email:     getEnvOrFlag("ISSUER_EMAIL", *issuerEmail),
				Phone:     getEnvOrFlag("ISSUER_PHONE", *issuerPhone),
			},
			Recipient:    receipient,
			Items:        items,
			TaxRate:      getEnvOrFlagFloat("TAX_RATE", *taxRate), // 6% tax
			CurrencyCode: getEnvOrFlag("CURRENCY_CODE", *currencyCode),
			Payment: models.PaymentInfo{
				Method:        getEnvOrFlag("PAYMENT_METHOD", *paymentMethod),
				BankName:      getEnvOrFlag("PAYMENT_BANK", *paymentBankName),
				AccountName:   getEnvOrFlag("PAYMENT_ACCOUNT_NAME", *paymentAccountName),
				AccountNumber: getEnvOrFlag("PAYMENT_ACCOUNT_NUMBER", *paymentAccountNumber),
			},
		}

		invID, err := repositories.InsertInvoice(tx, inv.Number, inv)
		if err != nil {
			http.Error(w, "Failed to save invoice: "+err.Error(), http.StatusInternalServerError)
			return
		}

		err = tx.Commit()
		if err != nil {
			http.Error(w, "Failed to commit transaction: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/invoice/%s", invID), http.StatusSeeOther)
	})

	// serve generated invoice
	http.HandleFunc("/invoice/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/invoice/"):]
		fmt.Printf("invoice view request from: %s\n, id: %s", r.RemoteAddr, id)
		inv, err := repositories.GetInvoice(db, id)
		if err != nil {
			http.Error(w, "Failed to get invoice: "+err.Error(), http.StatusInternalServerError)
			return
		}
		pages.InvoicePage(inv).Render(r.Context(), w)
	})

	fmt.Println("Server running on http://" + getEnvOrFlag("SERVE_ADDRESS", *serveAddress))
	http.ListenAndServe(getEnvOrFlag("SERVE_ADDRESS", *serveAddress), nil)
	fmt.Println("Server stopped")
}
