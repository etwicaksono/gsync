# Run wire to generate dependency injection code
wire:
	@echo "Running Wire..."
	wire ./injector

deploy:
	go build -o "D:\Apps\gsync\gsync.exe" "cmd/cli/main.go"
	copy .env "D:\Apps\gsync\.env"

run:
	go run ./cmd/cli/main.go

run-finance:
	go run ./cmd/cli/main.go finance

run-finance-pull:
	go run ./cmd/cli/main.go finance pull

run-finance-pull-h:
	go run ./cmd/cli/main.go finance pull -h

run-finance-pull-all:
	go run ./cmd/cli/main.go finance pull all

run-finance-pull-transaction:
	go run ./cmd/cli/main.go finance pull transaction

run-finance-pull-credit:
	go run ./cmd/cli/main.go finance pull credit

run-finance-pull-debt:
	go run ./cmd/cli/main.go finance pull debt

run-finance-pull-transfer:
	go run ./cmd/cli/main.go finance pull transfer

run-finance-pull-account:
	go run ./cmd/cli/main.go finance pull account

run-finance-pull-category:
	go run ./cmd/cli/main.go finance pull category

run-finance-push:
	go run ./cmd/cli/main.go finance push -c "Resume!B3" -s "2022-12-01" -e "2026-12-31"

run-finance-push-resume-detail-all:
	go run ./cmd/cli/main.go finance push resume-detail -c "Resume!B3" -s "2022-12-01" -e "2026-12-31"

run-finance-push-resume-detail-2023:
	go run ./cmd/cli/main.go finance push resume-detail -c "2023!B88" -s "2022-12-01" -e "2023-12-31"

run-finance-push-resume-detail-2024:
	go run ./cmd/cli/main.go finance push resume-detail -c "2024!B88" -s "2024-01-01" -e "2024-12-31"

run-finance-push-resume-detail-2025:
	go run ./cmd/cli/main.go finance push resume-detail -c "2025!B88" -s "2025-01-01" -e "2025-12-31"

run-finance-push-resume-detail-2026:
	go run ./cmd/cli/main.go finance push resume-detail -c "2026!B88" -s "2026-01-01" -e "2026-12-31"

run-finance-push-resume-detail-2027:
	go run ./cmd/cli/main.go finance push resume-detail -c "2027!B88" -s "2027-01-01" -e "2027-12-31"

run-finance-push-resume-recap-2023:
	go run ./cmd/cli/main.go finance push resume-recap -c "2023!B50" -s "2022-12-01" -e "2023-12-31"

run-finance-push-resume-recap-2024:
	go run ./cmd/cli/main.go finance push resume-recap -c "2024!B50" -s "2024-01-01" -e "2024-12-31"

run-finance-push-resume-recap-2025:
	go run ./cmd/cli/main.go finance push resume-recap -c "2025!B50" -s "2025-01-01" -e "2025-12-31"

run-finance-push-resume-recap-2026:
	go run ./cmd/cli/main.go finance push resume-recap -c "2026!B50" -s "2026-01-01" -e "2026-12-31"

run-finance-push-resume-recap-2027:
	go run ./cmd/cli/main.go finance push resume-recap -c "2027!B50" -s "2027-01-01" -e "2027-12-31"

run-finance-sync-2025:
	go run ./cmd/cli/main.go finance pull all
	go run ./cmd/cli/main.go finance push resume-detail -c "2025!B88" -s "2025-01-01" -e "2025-12-31"
	go run ./cmd/cli/main.go finance push resume-recap -c "2025!B50" -s "2025-01-01" -e "2025-12-31"

# API Push commands (push data from Google Sheet to Finance Manager API)
SKIP ?= 0
LIMIT ?= 0

run-finance-api-push:
	go run ./cmd/cli/main.go finance api-push

run-finance-api-push-transaction:
	go run ./cmd/cli/main.go finance api-push transaction -s $(SKIP) -l $(LIMIT)

run-finance-api-push-transfer:
	go run ./cmd/cli/main.go finance api-push transfer -s $(SKIP) -l $(LIMIT)
# ============================================

# ============================================
# ============================================
# These commands run the BudgetBakers browser automation scripts
# Located in: scripts/budgetbakers/
#
# First time setup:
#   make bb-setup
#
# Usage:
#   make bb-transaction-auto SKIP=0 LIMIT=10
#   make bb-transfer-auto SKIP=0 LIMIT=5
# ============================================

BUDGETBAKERS_DIR = scripts/budgetbakers

bb-setup:
	cd $(BUDGETBAKERS_DIR) && npm install && npm run setup

bb-transaction-auto:
	cd $(BUDGETBAKERS_DIR) && node budgetbakers-transaction.js --skip $(SKIP) --limit $(LIMIT) --mode auto

bb-transaction-assist:
	cd $(BUDGETBAKERS_DIR) && node budgetbakers-transaction.js --skip $(SKIP) --limit $(LIMIT) --mode assist

bb-transfer-auto:
	cd $(BUDGETBAKERS_DIR) && node budgetbakers-transfer.js --skip $(SKIP) --limit $(LIMIT) --mode auto

bb-transfer-assist:
	cd $(BUDGETBAKERS_DIR) && node budgetbakers-transfer.js --skip $(SKIP) --limit $(LIMIT) --mode assist

bb-clean:
	cd $(BUDGETBAKERS_DIR) && del /f /q .budgetbakers-session 2>nul || true
