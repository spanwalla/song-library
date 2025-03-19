include .env
export

compose-up: ### Run docker-compose
	docker-compose up --build -d && docker-compose logs -f
.PHONY: compose-up

compose-down: ### Down docker-compose
	docker-compose down --remove-orphans
.PHONY: compose-down

docker-rm-volume: ### Remove docker volume
	docker volume rm song-library_postgres_data
.PHONY: docker-rm-volume

migrate-create:  ### Create new migration
	migrate create -ext sql -dir migrations 'song_library'
.PHONY: migrate-create

migrate-up: ### Migration up
	migrate -path migrations -database 'postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:{$(POSTGRES_PORT)}/$(POSTGRES_DB)?sslmode=disable' up
.PHONY: migrate-up

migrate-down: ### Migration down
	echo "y" | migrate -path migrations -database 'postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:{$(POSTGRES_PORT)}/$(POSTGRES_DB)?sslmode=disable' down
.PHONY: migrate-down

linter-golangci: ### Check by golangci linter
	go tool github.com/golangci/golangci-lint/cmd/golangci-lint run
.PHONY: linter-golangci

swag: ### Generate swagger docs
	go tool github.com/swaggo/swag/cmd/swag init -g 'internal/app/app.go' --parseInternal --parseDependency
.PHONY: swag

test: ### Run test
	go test -v './internal/...'
.PHONY: test

mockgen: ### Generate mock
	go tool go.uber.org/mock/mockgen -source='internal/service/service.go'       -destination='internal/mocks/service/mock.go'    -package=servicemocks
	go tool go.uber.org/mock/mockgen -source='internal/repository/repository.go' -destination='internal/mocks/repository/mock.go' -package=repomocks
	go tool go.uber.org/mock/mockgen -source='internal/webapi/webapi.go'         -destination='internal/mocks/webapi/mock.go'     -package=webapimocks
.PHONY: mockgen

bin-deps: ### Install binary dependencies
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
.PHONY: bin-deps