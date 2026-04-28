# Format Go code using goimports
format:
	goimports -local github.com/akhmed9505/delayed-notifier -w .

# Run golangci-lint (main lint command)
lint:
	golangci-lint run ./...

# Run golangci-lint with auto-fix where possible
lint-fix:
	golangci-lint run ./... --fix

# Run all tests (unit + integration)
test:
	go test ./...

# Build and start all Docker services
docker-up:
	docker compose up --build

# Stop and remove all Docker services and volumes
docker-down:
	docker compose down -v
