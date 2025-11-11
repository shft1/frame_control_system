APP_NAME=frame_control_system

.PHONY: build run tidy

build:
	go build -o bin/$(APP_NAME) ./cmd/server

run:
	go run ./cmd/server

tidy:
	go mod tidy


