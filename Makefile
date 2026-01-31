APP_NAME := ne-noy
CMD_DIR := ./cmd/$(APP_NAME)
CONFIG_FILE := ./configs/config.yaml

.PHONY: build run test clean fmt vet migrate-up migrate-down doc

migrate-up:
	goose -dir ./migrations up

migrate-down:
	goose -dir ./migrations down

build: doc
	go build -o bin/$(APP_NAME) $(CMD_DIR)

run: build
	./bin/$(APP_NAME)

test:
	go test ./...

clean:
	go clean
	rm -rf bin

fmt:
	go fmt ./...
	swag fmt

vet:
	go vet ./...

doc:
	swag init -g main.go -d cmd/ne-noy,internal
