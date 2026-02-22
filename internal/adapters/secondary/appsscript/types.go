package appsscript

// SheetTransaction represents a transaction from Google Sheet
type SheetTransaction struct {
	No       string `json:"no"`
	Date     string `json:"date"` // DD/MM/YYYY
	Account  string `json:"account"`
	Category string `json:"category"`
	Amount   string `json:"amount"`
	Note     string `json:"note"`
	Type     string `json:"type"` // "expense" or "income"
}

// SheetTransfer represents a transfer from Google Sheet
type SheetTransfer struct {
	No          string `json:"no"`
	Date        string `json:"date"` // DD/MM/YYYY
	FromAccount string `json:"from_account"`
	ToAccount   string `json:"to_account"`
	Amount      string `json:"amount"`
	Note        string `json:"note"`
}

// SheetMeta contains pagination metadata
type SheetMeta struct {
	TotalData     int `json:"total_data"`
	TotalFiltered int `json:"total_filtered"`
	Skip          int `json:"skip"`
	Limit         int `json:"limit"`
}

// SheetTransactionResponse is the response for transaction fetch
type SheetTransactionResponse struct {
	Data []SheetTransaction `json:"data"`
	Meta SheetMeta          `json:"meta"`
}

// SheetTransferResponse is the response for transfer fetch
type SheetTransferResponse struct {
	Data []SheetTransfer `json:"data"`
	Meta SheetMeta       `json:"meta"`
}

// UpdateStatusRequest is the request to update status in Google Sheet
type UpdateStatusRequest struct {
	No     string `json:"no"`
	Status string `json:"status"`
	Type   string `json:"type"`   // "transaction" or "transfer"
	Target string `json:"target"` // "budgetbakers" or "financeapp"
}

// UpdateStatusResponse is the response from status update
type UpdateStatusResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	No      string `json:"no,omitempty"`
	Type    string `json:"type,omitempty"`
}
