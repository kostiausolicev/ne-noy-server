APP_NAME := ne-noy
CMD_DIR := ./cmd/$(APP_NAME)
CONFIG_FILE := ./configs/config.yaml
DOC_DESTINATION=docs
DOCKER_VERSION=0.1.3
PROTO_DIR := ./proto
PROTO_GEN_DIR := ./proto/gen

.PHONY: build run test clean fmt vet migrate-up migrate-down doc docker-build docker-save proto-gen

migrate-up:
	goose -dir ./migrations up

migrate-down:
	goose -dir ./migrations down

build: doc
	go build -o bin/$(APP_NAME) $(CMD_DIR)

run: build
	./bin/$(APP_NAME)

test: build
	@echo "Run test..."
	go test ./internal/service -v
	@echo "Test completed"

clean:
	go clean
	rm -rf bin

fmt:
	go fmt ./...
	swag fmt

vet:
	go vet ./...

doc:
	@rm -rf $(DOC_DESTINATION)
	swag init -g main.go -d cmd/ne-noy,internal

docker-build:
	docker buildx build --platform linux/amd64 -f ./build/package/Dockerfile -t $(APP_NAME):$(DOCKER_VERSION) --load .

docker-save: docker-build
	docker save $(APP_NAME):$(DOCKER_VERSION) -o $(APP_NAME)-$(DOCKER_VERSION).tar

proto-gen:
	@mkdir -p $(PROTO_GEN_DIR)
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(PROTO_GEN_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_GEN_DIR) \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/model.proto $(PROTO_DIR)/attachment_service.proto
