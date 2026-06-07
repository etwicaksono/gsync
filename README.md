# 🚀 Gsync (Google Sheet Manager)

![Go Version](https://img.shields.io/badge/Go-1.21-00ADD8?style=flat-square&logo=go)
![CLI](https://img.shields.io/badge/CLI-Cobra-red?style=flat-square)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)

**Gsync** is a powerful Command Line Interface (CLI) application built in Go that synchronizes your financial data between Google Sheets, a local PostgreSQL database, and third-party finance tracking apps like **BudgetBakers** and **BudgetMate**. 

Tired of doing manual data entry across multiple financial apps? Let Gsync do the heavy lifting for you! ✨

---

## ✨ Features

- 🔄 **Two-Way Synchronization**: Pull and push data between Google Sheets and your local database.
- 🏦 **BudgetBakers Integration**: Automatically sync your Google Sheet transactions and transfers to your BudgetBakers wallet via REST API.
- 💳 **BudgetMate Integration**: Push financial records seamlessly to your personal BudgetMate API instance.
- 📜 **Apps Script Webhook**: Communicates directly with Google Apps Script to read ranges and update cell statuses (e.g., `INSERTED TO BudgetMate`) without manual Google Sheets API overhead.
- 🚀 **Hexagonal Architecture**: Clean, modular, and testable codebase using Ports & Adapters.
- ⚡ **Batch Processing**: Smart rate-limiting and batched API requests to ensure stability and efficiency.

---

## 🛠️ Prerequisites

Before you begin, ensure you have met the following requirements:
- **Go** >= 1.21 installed
- **PostgreSQL** running locally or remotely
- A **Google Sheet** configured with the appropriate tabs (`Transaction`, `Transfer`, `Categories`, etc.)
- **Google Apps Script** deployed as a Web App (use the script provided in `scripts/budgetbakers/google-apps-script.js`)
- Credentials for your target platforms (BudgetBakers API Token / BudgetMate account)

---

## 🚀 Installation & Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/yourusername/gsync.git
   cd gsync
   ```

2. **Configure Environment Variables:**
   Copy the example config file and fill in your actual credentials.
   ```bash
   cp .env.example .env
   ```
   > 💡 *Check out `.env.example` for detailed comments on what each variable does.*

3. **Deploy the Apps Script:**
   - Copy the code from `scripts/budgetbakers/google-apps-script.js`.
   - Go to your Google Sheet -> `Extensions` -> `Apps Script`.
   - Paste the code, save, and hit **Deploy > New Deployment** as a Web App.
   - Copy the resulting Web App URL and paste it into `APPS_SCRIPT_URL` in your `.env`.

4. **Build the CLI:**
   ```bash
   go build -o gsync.exe cmd/cli/main.go
   ```

---

## 💻 Usage

Run the CLI using the compiled binary (or via `go run`).

```bash
# View all available commands
./gsync --help
```

### 📊 1. GSheet Finance Sync
Synchronize core financial records between your local database and Google Sheets.
```bash
./gsync gsheet-finance pull     # Pulls data from GSheet into Database
./gsync gsheet-finance push     # Pushes summary data from DB to GSheet
```

### 💸 2. BudgetBakers Integration
Push transactions directly to BudgetBakers.
```bash
# Push transactions
./gsync budgetbakers transaction

# Push transfers
./gsync budgetbakers transfer

# Debug accounts and category mappings
./gsync budgetbakers debug
```
> *Use the `--skip` and `--limit` flags to control how many rows are processed at once.*

### 👛 3. BudgetMate Integration
Push transactions to your custom BudgetMate API.
```bash
./gsync budgetmate transaction
./gsync budgetmate transfer
```

---

## 🏗️ Architecture overview

Gsync follows **Hexagonal Architecture (Ports and Adapters)**:
- **`cmd/cli/`**: Application entrypoint.
- **`internal/commands/`**: Cobra CLI definitions and flags.
- **`internal/adapters/app/`**: Core business logic and orchestration (e.g., `budgetmate_app`, `budgetbakers_app`, `gsheet_finance_app`).
- **`internal/adapters/secondary/`**: Infrastructure and external integrations (e.g., BudgetBakers REST, Google Apps Script HTTP clients, GORM).
- **`internal/ports/`**: Interfaces defining the boundaries of the application.

---

## 📝 License

Distributed under the MIT License. See `LICENSE` for more information.

---
*Built with ❤️ to keep your finances perfectly balanced.*
