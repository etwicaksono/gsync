 package financeapi
 
 import "time"
 
 // LoginRequest is the request body for authentication
 type LoginRequest struct {
 	Email    string `json:"email"`
 	Password string `json:"password"`
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
 
 // Category represents a category from Finance Manager
 type Category struct {
 	ID       string `json:"id"`
 	Name     string `json:"name"`
 	Type     string `json:"type,omitempty"` // "expense" or "income"
 	ParentID string `json:"parent_id,omitempty"`
 	Icon     string `json:"icon,omitempty"`
 }
 
 // CreateTransactionRequest is the request to create a transaction
 type CreateTransactionRequest struct {
 	Amount        float64   `json:"amount"`
 	AccountID     string    `json:"account_id"`
 	CategoryID    string    `json:"category_id"`
 	DateTime      time.Time `json:"date_time"`
 	Type          string    `json:"type"` // "expense" or "income"
 	PaymentStatus string    `json:"payment_status,omitempty"`
 	Description   string    `json:"description,omitempty"`
 	Labels        []string  `json:"labels,omitempty"`
 	Payee         string    `json:"payee,omitempty"`
 	PaymentMethod string    `json:"payment_method,omitempty"`
 }
 
 // CreateTransactionResponse is the response from creating a transaction
 type CreateTransactionResponse struct {
 	ID      string `json:"id,omitempty"`
 	Message string `json:"message,omitempty"`
 	Error   string `json:"error,omitempty"`
 }
 
 // CreateTransferRequest is the request to create a transfer
 type CreateTransferRequest struct {
 	Amount         float64   `json:"amount"`
 	FromAccountID  string    `json:"from_account_id"`
 	ToAccountID    string    `json:"to_account_id"`
 	AmountReceived float64   `json:"amount_received,omitempty"`
 	DateTime       time.Time `json:"date_time"`
 	PaymentStatus  string    `json:"payment_status,omitempty"`
 	Description    string    `json:"description,omitempty"`
 	Labels         []string  `json:"labels,omitempty"`
 	PaymentMethod  string    `json:"payment_method,omitempty"`
 }
 
 // CreateTransferResponse is the response from creating a transfer
 type CreateTransferResponse struct {
 	ID      string `json:"id,omitempty"`
 	Message string `json:"message,omitempty"`
 	Error   string `json:"error,omitempty"`
 }
