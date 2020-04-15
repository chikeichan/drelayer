ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
COMMIT := $(shell git log -1 --format='%H')

ldflags = -X ddrp-relayer/version.GitCommit=$(COMMIT)

BUILD_FLAGS := -ldflags '$(ldflags)'

create-migration:
	migrate create -ext sql -dir ./store/migrations -seq $(NAME)

migrate-local:
	migrate -source file://./store/migrations -database postgres://localhost/ddrp_relayer_development?sslmode=disable -verbose up

migrate-local-reset:
	migrate -source file://./store/migrations -database postgres://localhost/ddrp_relayer_development?sslmode=disable -verbose drop
	migrate -source file://./store/migrations -database postgres://localhost/ddrp_relayer_development?sslmode=disable -verbose up

fmt:
	gofmt -s -w .
	goimports -w .

swagger:
	rm -rf ./restmodels/* && swagger generate model -f ./swagger/swagger.yml -m restmodels

clean:
	rm -rf ./build

drelayer:
	go build $(BUILD_FLAGS) -o ./build/drelayer ./cmd/main.go

all-cross:
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o ./build/drelayer-darwin-x64 ./cmd/main.go
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o ./build/drelayer-linux-x64 ./cmd/main.go

test:
	MIGRATIONS_DIR=$(ROOT_DIR)/store/migrations go test -v --race ./...

.PHONY: create-migration fmt swagger drelayer test clean all-cross