package budgetmateapi

// LoginRequest is the request body for authentication
 type LoginRequest struct {
	EmailOrUsername string `json:"email_or_username"`
	Password        string `json:"password"`
 }
 
 // LoginResponse is the response from authentication
 type LoginResponse struct {
 	Token   string `json:"token,omitempty"`
 	Message string `json:"message,omitempty"`
 	Error   string `json:"error,omitempty"`
 }
 
 // Account represents an account from Finance Manager
 type Account struct {
 	ID       string  `json:"id"`
 	Name     string  `json:"name"`
 	Type     string  `json:"type,omitempty"`
 	Currency string  `json:"currency,omitempty"`
 	Balance  float64 `json:"balance,omitempty"`
 	IsActive bool    `json:"is_active,omitempty"`
 }

 // AccountsResponse wraps the accounts API response
type AccountsResponse struct {
	Success bool      `json:"success"`
	Data    []Account `json:"data"`
}
 
 // Category represents a category from Finance Manager
 type Category struct {
 	ID       string `json:"id"`
 	Name     string `json:"name"`
 	Type     string `json:"type,omitempty"` // "expense" or "income"
 	ParentID string `json:"parent_id,omitempty"`
 	Icon     string `json:"icon,omitempty"`
 }

 // CategoriesResponse wraps the categories API response
type CategoriesResponse struct {
	Success bool       `json:"success"`
	Data    []Category `json:"data"`
}

// Label represents a label from Finance Manager
type Label struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

// LabelsResponse wraps the labels API response
type LabelsResponse struct {
	Success bool    `json:"success"`
	Data    []Label `json:"data"`
}
 
 // CreateTransactionRequest is the request to create a transaction
 type CreateTransactionRequest struct {
 	Amount        float64   `json:"amount"`
 	AccountID     string    `json:"account_id"`
 	CategoryID    string    `json:"category_id"`
	Date          string    `json:"date"`
 	Type          string    `json:"type"` // "expense" or "income"
 	PaymentStatus string    `json:"payment_status,omitempty"`
 	Description   string    `json:"description,omitempty"`
	Labels        []string  `json:"label_ids,omitempty"`
 	Payee         string    `json:"payee,omitempty"`
 	PaymentMethod string    `json:"payment_method,omitempty"`
 }
 
 // CreateTransactionResponse is the response from creating a transaction
 type CreateTransactionResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
	Data    struct {
		ID string `json:"id,omitempty"`
		// Labels echoes back what was actually stored. Comparing this against
		// the request is the only way to catch a payload the API accepts but
		// silently drops.
		Labels []Label `json:"labels,omitempty"`
	} `json:"data,omitempty"`
	Error interface{} `json:"error,omitempty"`
 }
 
 // CreateTransferRequest is the request to create a transfer
 type CreateTransferRequest struct {
 	Amount         float64   `json:"amount"`
 	FromAccountID  string    `json:"from_account_id"`
 	ToAccountID    string    `json:"to_account_id"`
 	AmountReceived float64   `json:"amount_received,omitempty"`
	Date           string    `json:"date"`
 	PaymentStatus  string    `json:"payment_status,omitempty"`
 	Description    string    `json:"description,omitempty"`
	// CreateTransferRequest has no label field in the API schema, so this is
	// accepted and then ignored. Kept only so existing callers still compile.
	Labels         []string  `json:"label_ids,omitempty"`
 	PaymentMethod  string    `json:"payment_method,omitempty"`
 }
 
 // CreateTransferResponse is the response from creating a transfer
 type CreateTransferResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
	Data    struct {
		ID string `json:"id,omitempty"`
	} `json:"data,omitempty"`
	Error interface{} `json:"error,omitempty"`
 }
