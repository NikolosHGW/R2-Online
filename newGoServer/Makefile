.PHONY: login game memscanner docker-up docker-down migrate tidy

login:
	go run ./cmd/login

game:
	go run ./cmd/game

# memscanner: Windows-only RC4 memory scanner.
# Usage: make memscanner PROC=R2ClientRU.exe
#        make memscanner PROC=R2ClientRU.exe DUMP=0x7FF600010000
PROC ?= R2ClientRU.exe
DUMP ?=
memscanner:
	go run ./cmd/memscanner -proc $(PROC) $(if $(DUMP),-dump $(DUMP),)

build:
	go build ./cmd/login
	go build ./cmd/game

docker-up:
	docker compose up -d postgres redis

docker-down:
	docker compose down

docker-all:
	docker compose up --build

migrate:
	@echo "Applying migrations..."
	docker compose exec postgres psql -U r2 -d r2online -f /docker-entrypoint-initdb.d/001_init.sql

tidy:
	go mod tidy

test:
	go test ./...
