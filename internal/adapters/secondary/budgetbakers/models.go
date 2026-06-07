package budgetbakers

import "time"



// AmountWithCurrency represents a monetary amount with optional currency code
type AmountWithCurrency struct {
	CurrencyCode string  `json:"currencyCode,omitempty"`
	Value        float64 `json:"value"`
}

// Account represents a BudgetBakers account
type Account struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	AccountType    string              `json:"accountType,omitempty"`
	CurrencyCode   string              `json:"currencyCode,omitempty"`
	InitialBalance *AmountWithCurrency `json:"initialBalance,omitempty"`
	Archived       bool                `json:"archived,omitempty"`
}

// Category represents a BudgetBakers category
type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Record represents a BudgetBakers transaction record payload
type Record struct {
	AccountID   string             `json:"accountId"`
	CategoryID  string             `json:"categoryId,omitempty"`
	Amount      AmountWithCurrency `json:"amount"`
	PaymentType string             `json:"paymentType"` // "cash", "debit_card", "credit_card", "transfer", "voucher", "mobile_payment", "web_payment"
	RecordState string             `json:"recordState,omitempty"` // "cleared", "reconciled", "uncleared"
	RecordDate  time.Time          `json:"recordDate"`
	Note        string             `json:"note,omitempty"`
}

// BatchRecordResponse represents the 207 Multi-Status response from POST /records
type BatchRecordResponse struct {
	Summary struct {
		Total        int `json:"total"`
		Succeeded    int `json:"succeeded"`
		ClientErrors int `json:"clientErrors"`
		ServerErrors int `json:"serverErrors"`
	} `json:"summary"`
	Results []struct {
		Success   bool   `json:"success"`
		Error     string `json:"error,omitempty"`
		ErrorType string `json:"errorType,omitempty"`
		ID        string `json:"id,omitempty"`
	} `json:"results"`
}
