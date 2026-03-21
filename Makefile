# Ivy - Development Makefile

# Docker
up:
	docker compose up -d

down:
	docker compose down

restart:
	docker compose restart backend

logs:
	docker compose logs -f backend

logs-all:
	docker compose logs -f

# Database
migrate-up:
	docker compose exec backend migrate -path /app/migrations -database "postgres://${DB_USER}:${DB_PASSWORD}@postgres:5432/${DB_NAME}?sslmode=disable" up

migrate-down-1:
	docker compose exec backend migrate -path /app/migrations -database "postgres://${DB_USER}:${DB_PASSWORD}@postgres:5432/${DB_NAME}?sslmode=disable" down 1

migrate-version:
	docker compose exec backend migrate -path /app/migrations -database "postgres://${DB_USER}:${DB_PASSWORD}@postgres:5432/${DB_NAME}?sslmode=disable" version

db-shell:
	docker compose exec postgres psql -U ${DB_USER} -d ${DB_NAME}

# Backend
backend-test:
	cd backend && go test ./... -v

backend-build:
	cd backend && go build -o bin/server ./cmd/server

e2e-test:
	cd backend && go test -tags e2e ./test/e2e/... -v -count=1

e2e-test-frontend:
	cd frontend && npx playwright test --reporter=list

e2e-test-all: e2e-test e2e-test-frontend

# Frontend
frontend-dev:
	cd frontend && npm run dev

frontend-install:
	cd frontend && npm install

frontend-test:
	cd frontend && npm test

# Clean
clean:
	docker compose down -v
	rm -rf backend/bin
	rm -rf frontend/dist
