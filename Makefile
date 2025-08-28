run:
	go run ./cmd/cli/main.go

deploy:
	go build -o "D:\Apps\gsync\gsync.exe" "cmd/cli/main.go"
	copy .env "D:\Apps\gsync\.env"

run-pull:
	go run ./cmd/cli/main.go pull

run-pull-h:
	go run ./cmd/cli/main.go pull -h

run-pull-all:
	go run ./cmd/cli/main.go pull all

run-pull-transaction:
	go run ./cmd/cli/main.go pull transaction

run-pull-credit:
	go run ./cmd/cli/main.go pull credit

run-pull-debt:
	go run ./cmd/cli/main.go pull debt

run-pull-transfer:
	go run ./cmd/cli/main.go pull transfer

run-pull-account:
	go run ./cmd/cli/main.go pull account

run-pull-category:
	go run ./cmd/cli/main.go pull category