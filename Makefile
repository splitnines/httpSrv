.DEFAULT_GOAL := build

DIR := $(HOME)/.local/bin
FILE_NAME := httpSrv

.PHONY:fmt vet build

clean:
	go clean
fmt: clean
	go fmt ./...
vet: fmt
	go vet ./...
build: vet
	go build -o $(FILE_NAME)
	cp $(FILE_NAME) $(DIR)
