 # API Input Implementation Plan
 
 ## Overview
 
 This document outlines the plan to add REST API-based transaction and transfer input functionality to the Gsync application, similar to the existing Playwright-based automation (`make transaction-auto`, `make transfer-auto`) but using direct HTTP API calls to `http://localhost:3000/`.
 
## Data Source: Google Apps Script

The existing Playwright scripts fetch transaction/transfer data from a **Google Apps Script Web App** that reads from Google Sheets.

### Apps Script API

**Base URL:** Configured in `reference/config.js` as `APPS_SCRIPT_URL`

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `?type=transaction` | Get transactions from "Transaction" sheet |
| GET | `?type=transfer` | Get transfers from "Transfer" sheet |
| GET | `?skip=N&limit=M` | Pagination support |
| POST | (body: `{no, status, type}`) | Update status after insert |

### Transaction Sheet Format (starting from B5)

| Column | Field | Description |
|--------|-------|-------------|
| B | No | Row identifier |
| C | Date | DD/MM/YYYY format |
| D | Account | Account name (e.g., "Cash Eko", "OVO Dewi") |
| E | Category | Category name (e.g., "Groceries", "Salary") |
| F | Amount | Positive number |
| G | Note | Transaction description |
| H | Type | "income" or "expense" |
| I | Status | Updated to "INSERTED" after processing |

### Transfer Sheet Format (starting from B5)

| Column | Field | Description |
|--------|-------|-------------|
| B | No | Row identifier |
| C | Date | DD/MM/YYYY format |
| D | From Account | Source account name |
| E | To Account | Destination account name |
| F | Amount | Transfer amount |
| G | Note | Transfer description |
| H | Status | Updated to "INSERTED" after processing |

### Response Format

```json
{
  "data": [
    {
      "no": "123",
      "date": "01/02/2026",
      "account": "Cash",
      "category": "Groceries",
      "amount": "50000",
      "note": "Weekly groceries",
      "type": "expense"
    }
  ],
  "meta": {
    "total_data": 150,
    "total_filtered": 50,
    "skip": 0,
    "limit": 50
  }
}
```

 ## Recorded API Endpoints
 
 Based on Playwright network recording, the Finance Manager application uses the following REST API endpoints:
 
 ### Authentication
 
 | Method | Endpoint | Description |
 |--------|----------|-------------|
 | POST | `/api/v1/auth/login` | Login with email/password, returns session token |
 
 ### Master Data (Read)
 
 | Method | Endpoint | Description |
 |--------|----------|-------------|
 | GET | `/api/v1/accounts` | Get all accounts |
 | GET | `/api/v1/accounts?is_active=true` | Get active accounts only |
 | GET | `/api/v1/categories` | Get all categories |
 | GET | `/api/v1/categories/tree` | Get categories in tree structure |
 | GET | `/api/v1/labels` | Get all labels |
 
 ### Transactions
 
 | Method | Endpoint | Description |
 |--------|----------|-------------|
 | GET | `/api/v1/transactions` | List transactions (with filters) |
 | POST | `/api/v1/transactions` | Create new transaction (expense/income) |
 
 ### Transfers
 
 | Method | Endpoint | Description |
 |--------|----------|-------------|
 | POST | `/api/v1/transfers` | Create new transfer between accounts |
 
 ## Form Fields Analysis
 
 ### Transaction Form (Expense/Income)
 
 | Field | Required | Description |
 |-------|----------|-------------|
 | Amount | Yes | Transaction amount (positive number) |
 | Account | Yes | Account ID (from `/api/v1/accounts`) |
 | Category | Yes | Category ID (from `/api/v1/categories`) |
 | Date & Time | Yes | ISO 8601 datetime format |
 | Payment Status | No | `Cleared`, `Pending`, `Scheduled` (default: Cleared) |
 | Description | No | Transaction note/description |
 | Labels | No | Array of label IDs |
 | Payee | No | Who was paid |
 | Payment Method | No | `Cash`, `Credit Card`, `Debit Card`, `Bank Transfer`, `Digital Wallet` |
 | Type | Yes | `expense` or `income` |
 
 ### Transfer Form
 
 | Field | Required | Description |
 |-------|----------|-------------|
 | Amount | Yes | Transfer amount |
 | From Account | Yes | Source account ID |
 | To Account | Yes | Destination account ID |
 | Amount Received | No | For multi-currency transfers (default: same as Amount) |
 | Date & Time | Yes | ISO 8601 datetime format |
 | Payment Status | No | `Cleared`, `Pending`, `Scheduled` (default: Cleared) |
 | Description | No | Transfer note |
 | Labels | No | Array of label IDs |
 | Payment Method | No | Payment method |
 
 ## Proposed Implementation
 
 ### New CLI Commands
 
 Add new subcommands under `finance` command:
 
 ```
 gsync finance api-push transaction [flags]
 gsync finance api-push transfer [flags]
 ```
 
 **Flags:**
 - `--skip <n>` - Number of records to skip (default: 0)
 - `--limit <n>` - Number of records to process, 0 = all (default: 0)
 - `--dry-run` - Preview without actually sending (optional)
 
 ### New Makefile Targets
 
 ```makefile
 run-finance-api-push-transaction:
 	go run ./cmd/cli/main.go finance api-push transaction --skip $(SKIP) --limit $(LIMIT)
 
 run-finance-api-push-transfer:
 	go run ./cmd/cli/main.go finance api-push transfer --skip $(SKIP) --limit $(LIMIT)
 ```
 
 ### File Structure
 
 ```
 internal/
 ├── adapters/
 │   └── secondary/
│       ├── gsheet/
│       │   └── apps_script/
│       │       ├── client.go         # HTTP client for Apps Script
│       │       ├── transaction.go    # Fetch transactions
│       │       ├── transfer.go       # Fetch transfers
│       │       ├── status.go         # Update status after insert
│       │       └── types.go          # Request/response structs
 │       └── api/
 │           └── finance_manager/
 │               ├── client.go         # HTTP client with auth
 │               ├── auth.go           # Login/session handling
 │               ├── transaction.go    # Transaction API calls
 │               ├── transfer.go       # Transfer API calls
 │               ├── account.go        # Account lookup
 │               ├── category.go       # Category lookup
 │               └── types.go          # Request/response structs
 ├── commands/
 │   └── finance_api_push.go           # CLI command handlers
 └── ports/
     └── secondary/
        ├── api/
        │   └── finance_manager.go    # Finance Manager API interface
        └── gsheet/
            └── apps_script.go        # Apps Script interface
 ```
 
 ### Configuration
 
 Add to `.env`:
 
 ```env
# Google Apps Script URL (for fetching data from Google Sheets)
APPS_SCRIPT_URL=https://script.google.com/macros/s/xxx/exec

# Finance Manager API (for pushing data)
 FINANCE_API_URL=http://localhost:3000
 FINANCE_API_EMAIL=your-email@example.com
 FINANCE_API_PASSWORD=your-password
 ```
 
 Add to `config.go`:
 
 ```go
type AppsScriptConfig struct {
    URL string
}

 type FinanceApiConfig struct {
     URL      string
     Email    string
     Password string
 }
 ```
 
### Apps Script Types

```go
// Transaction from Google Sheet
type SheetTransaction struct {
    No       string `json:"no"`
    Date     string `json:"date"`      // DD/MM/YYYY
    Account  string `json:"account"`
    Category string `json:"category"`
    Amount   string `json:"amount"`
    Note     string `json:"note"`
    Type     string `json:"type"`      // "expense" or "income"
}

// Transfer from Google Sheet
type SheetTransfer struct {
    No          string `json:"no"`
    Date        string `json:"date"`         // DD/MM/YYYY
    FromAccount string `json:"from_account"`
    ToAccount   string `json:"to_account"`
    Amount      string `json:"amount"`
    Note        string `json:"note"`
}

// Response wrapper
type SheetResponse[T any] struct {
    Data []T `json:"data"`
    Meta struct {
        TotalData     int `json:"total_data"`
        TotalFiltered int `json:"total_filtered"`
        Skip          int `json:"skip"`
        Limit         int `json:"limit"`
    } `json:"meta"`
}

// Status update request
type UpdateStatusRequest struct {
    No     string `json:"no"`
    Status string `json:"status"`
    Type   string `json:"type"` // "transaction" or "transfer"
}
```

 ### API Request/Response Types
 
 #### Login Request
 ```go
 type LoginRequest struct {
     Email    string `json:"email"`
     Password string `json:"password"`
 }
 
 type LoginResponse struct {
     Token string `json:"token"`
     // or session cookie
 }
 ```
 
 #### Transaction Request
 ```go
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
 ```
 
 #### Transfer Request
 ```go
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
 ```
 
 ### Data Flow
 
```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────────┐
│  Google Sheet   │────▶│  Apps Script     │────▶│  Gsync CLI          │
│  (Transaction/  │     │  Web App         │     │  (Go Application)   │
│   Transfer)     │     │  (GET data)      │     │                     │
└─────────────────┘     └──────────────────┘     └──────────┬──────────┘
                                                            │
                        ┌──────────────────┐                │
                        │  Apps Script     │◀───────────────┤
                        │  Web App         │  (POST status) │
                        │  (Update status) │                │
                        └──────────────────┘                │
                                                            ▼
                        ┌──────────────────┐     ┌─────────────────────┐
                        │  Finance Manager │◀────│  Create Transaction │
                        │  API             │     │  or Transfer        │
                        │  localhost:3000  │     │                     │
                        └──────────────────┘     └─────────────────────┘
```

### Detailed Steps

1. **Fetch data from Google Sheets** via Apps Script URL
   - GET `{APPS_SCRIPT_URL}?type=transaction&skip=0&limit=50`
   - GET `{APPS_SCRIPT_URL}?type=transfer&skip=0&limit=50`

2. **Login to Finance Manager API** (once at start)
   - POST `/api/v1/auth/login` with email/password
   - Store auth token/cookie for subsequent requests

3. **Fetch master data** (accounts, categories) for name-to-ID mapping
   - GET `/api/v1/accounts`
   - GET `/api/v1/categories`
   - Build lookup maps: `accountName -> accountID`, `categoryName -> categoryID`

4. **For each record:**
    - Map account/category names to IDs
   - Convert date from DD/MM/YYYY to ISO 8601
   - Send POST request to Finance Manager API
   - On success: Update status in Google Sheet via Apps Script POST
   - Log progress

5. **Report summary** (success/fail counts)
 
 ### Account/Category Mapping
 
 The existing Playwright scripts use account/category **names** (e.g., "Cash", "Groceries").
 The API requires **IDs**. We need to:
 
 1. Fetch all accounts from `/api/v1/accounts`
 2. Fetch all categories from `/api/v1/categories`
 3. Build name-to-ID lookup maps
 4. Handle case-insensitive matching
 5. Handle "not found" scenarios gracefully
 
 ### Error Handling
 
 - **Auth errors (401)**: Re-login and retry
 - **Validation errors (400)**: Log error, continue to next record
 - **Server errors (5xx)**: Retry with exponential backoff
 - **Network errors**: Retry with limit
 
 ## Implementation Steps
 
1. **Phase 1: Apps Script Client**
   - Create HTTP client for Apps Script
   - Implement FetchTransactions
   - Implement FetchTransfers
   - Implement UpdateStatus

2. **Phase 2: Finance Manager API Client**
    - Create HTTP client with auth support
    - Implement login flow
    - Add retry/error handling
 
3. **Phase 3: Master Data**
    - Implement account fetching
    - Implement category fetching
    - Build name-to-ID lookup maps
 
4. **Phase 4: Transaction API**
    - Implement CreateTransaction
    - Add CLI command `finance api-push transaction`
    - Test with sample data
 
5. **Phase 5: Transfer API**
    - Implement CreateTransfer
    - Add CLI command `finance api-push transfer`
    - Test with sample data
 
6. **Phase 6: Integration**
    - Add Makefile targets
    - Update documentation
    - End-to-end testing
 
 ## Estimated Effort
 
 | Phase | Estimated Time |
 |-------|----------------|
| Phase 1: Apps Script Client | 1 hour |
| Phase 2: Finance Manager API Client | 1-2 hours |
| Phase 3: Master Data | 1 hour |
| Phase 4: Transaction API | 1-2 hours |
| Phase 5: Transfer API | 1 hour |
| Phase 6: Integration | 1 hour |
| **Total** | **6-8 hours** |
 
 ## Comparison: Playwright vs REST API
 
 | Aspect | Playwright | REST API |
 |--------|------------|----------|
 | Speed | Slower (browser rendering) | Faster (direct HTTP) |
 | Reliability | Less stable (UI changes) | More stable (API contract) |
 | Resource Usage | High (browser process) | Low (HTTP only) |
 | Dependencies | Node.js, Playwright | None (pure Go) |
 | Error Handling | Complex (DOM state) | Simple (HTTP status) |
 | Session | Browser cookies | JWT/Cookie token |
 
 ## Notes
 
 - The API response format needs to be verified by inspecting actual responses
 - Authentication mechanism (JWT vs session cookie) needs confirmation
 - Field names in API request body may differ from form field names
 - Consider adding a `--verbose` flag for debugging API calls
 
 ## Approval Required
 
 Before proceeding with implementation, please review and approve this plan.
