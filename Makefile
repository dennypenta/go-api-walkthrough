gen:
	go generate ./...

install:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.1

migrate_new:
	migrate create -ext sql -dir migrations -seq -digits 4 ${MNAME}
