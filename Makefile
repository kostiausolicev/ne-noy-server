APP_NAME := ne-noy
CMD_DIR := ./cmd/$(APP_NAME)
CONFIG_FILE := ./configs/config.yaml
MOCKS_DESTINATION=tests/mocks
DOC_DESTINATION=docs

.PHONY: build run test clean fmt vet migrate-up migrate-down doc mocks

migrate-up:
	goose -dir ./migrations up

migrate-down:
	goose -dir ./migrations down

build: doc
	go build -o bin/$(APP_NAME) $(CMD_DIR)

run: build
	./bin/$(APP_NAME)

mocks:
	@echo "Generate mocks..."
	@which mockgen
	@rm -rf $(MOCKS_DESTINATION)
	mockgen -source=internal/service/user_service.go -destination=$(MOCKS_DESTINATION)/mock_user_service.go -package=mocks
	mockgen -source=internal/service/event_service.go -destination=$(MOCKS_DESTINATION)/mock_event_service.go -package=mocks
	mockgen -source=internal/service/event_participant_service.go -destination=$(MOCKS_DESTINATION)/mock_event_participant_service.go -package=mocks
	@echo "Service mocks generated"
	mockgen -source=internal/repository/user_repository.go -destination=$(MOCKS_DESTINATION)/mock_user_repository.go -package=mocks
	mockgen -source=internal/repository/event_repository.go -destination=$(MOCKS_DESTINATION)/mock_event_repository.go -package=mocks
	mockgen -source=internal/repository/event_participant_repository.go -destination=$(MOCKS_DESTINATION)/mock_event_participant_repository.go -package=mocks
	mockgen -source=internal/repository/role_repository.go -destination=$(MOCKS_DESTINATION)/mock_role_repository.go -package=mocks
	@echo "Generate mocks completed"

test: build mocks
	@echo "Run test..."
	go test ./internal -v
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
