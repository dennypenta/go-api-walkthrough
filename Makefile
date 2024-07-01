DB_DSN ?= "postgres://pguser:pgpass@localhost:5432/main?sslmode=disable"

gen:
	go generate ./...

install:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.1
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1
	go install golang.org/x/tools/cmd/goimports@v0.22.0

migrate_new:
	migrate create -ext sql -dir migrations -seq -digits 4 ${MNAME}

migrate_up:
	migrate -path migrations -database ${DB_DSN} up

migrate_down:
	migrate -path migrations -database ${DB_DSN} down

migrate_fix:
	migrate -path migrations -database ${DB_DSN} force ${V}

migrate_v:
	migrate -path migrations -database ${DB_DSN} version

test_unit:
	go test -race -count 1 ./... -v -coverprofile=coverage.out 

test_integration:
	docker compose -f docker-compose.e2e.yaml up --build -d
	
	@echo "Checking e2e test environment is running..."
	until $$(curl --output /dev/null --silent --fail http://localhost:8080/healthz); do printf '.'; sleep 1; done && echo "Service Ready!"
	@echo 'Service has been started'
	
	go test -tags integration -race -count 1 ./tests -v
	
	docker compose -f docker-compose.e2e.yaml down

lint:
	goimports -l -w . && golangci-lint run
