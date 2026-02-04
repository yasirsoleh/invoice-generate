package models

type PaymentInfo struct {
	Method        string `json:"method"`
	BankName      string `json:"bank_name"`
	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number"`
	Notes         string `json:"notes,omitempty"`
}

type Party struct {
	Name      string `json:"name"`
	NRIC      string `json:"nric,omitempty"`
	BRN       string `json:"brn,omitempty"`
	TaxNumber string `json:"tax_number,omitempty"`
	Address   string `json:"address,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Email     string `json:"email,omitempty"`
}

type Item struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type Invoice struct {
	Number    string  `json:"number"`
	Date      string  `json:"date"`
	Issuer    Party   `json:"issuer"`
	Recipient Party   `json:"recipient"`
	Items     []Item  `json:"items"`
	TaxRate   float64 `json:"tax_rate"` // e.g. 0.06 for 6%
	Payment   PaymentInfo
}

func NewInvoice(number, date string, issuer, recipient Party, items []Item) Invoice {
	return Invoice{
		Number:    number,
		Date:      date,
		Issuer:    issuer,
		Recipient: recipient,
		Items:     items,
	}
}

func (i Invoice) Subtotal() float64 {
	total := 0.0
	for _, it := range i.Items {
		total += float64(it.Quantity) * it.Price
	}
	return total
}

func (i Invoice) ItemTotal(it Item) float64 {
	return float64(it.Quantity) * it.Price
}

func (i Invoice) SalesTax() float64 {
	return i.Subtotal() * i.TaxRate
}

func (i Invoice) TotalDue() float64 {
	return i.Subtotal() + i.SalesTax()
}
