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
 
 // CreateTransactionRequest is the request to create a transaction
 type CreateTransactionRequest struct {
 	Amount        float64   `json:"amount"`
 	AccountID     string    `json:"account_id"`
 	CategoryID    string    `json:"category_id"`
	Date          string    `json:"date"`
 	Type          string    `json:"type"` // "expense" or "income"
 	PaymentStatus string    `json:"payment_status,omitempty"`
 	Description   string    `json:"description,omitempty"`
 	Labels        []string  `json:"labels,omitempty"`
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
 	Labels         []string  `json:"labels,omitempty"`
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
