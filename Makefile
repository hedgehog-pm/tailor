.PHONY: fmt lint generate build

fmt:
	go tool golangci-lint fmt

lint:
	go tool golangci-lint config verify
	go tool golangci-lint run

generate:
	go tool oapi-codegen -config api/oapi-codegen.yaml api/task-schemas.yaml

build:
	go build -o bin/tailor ./cmd

.PHONY: fix
fix:
	go tool golangci-lint config verify
	go tool golangci-lint run --fix ./...
