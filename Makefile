all: init build

BINARY_NAME := pg-press

SERVER_ADDR := :9020
SERVER_PATH_PREFIX := /${BINARY_NAME}
#SERVER_PATH_PREFIX :=

clean:
	git clean -xfd

init:
	go mod tidy -v
	git submodule init
	git submodule update --recursive

run:
	SERVER_PATH_PREFIX=${SERVER_PATH_PREFIX} \
		go run ./cmd/${BINARY_NAME} server -a ${SERVER_ADDR}

dev:
	which gow || (echo 'gow is not installed, install with: `go install github.com/mitranim/gow@latest`' && exit 1)
	SERVER_PATH_PREFIX=${SERVER_PATH_PREFIX} \
		gow -e=go,json,html,js,css -r run ./cmd/${BINARY_NAME} server --addr ${SERVER_ADDR}

build:
	go build -v -o ./bin/${BINARY_NAME} ./cmd/${BINARY_NAME}

count:
	find . | grep -e '\.go$$' -e '\.html$$' -e '\.css$$' -e '\.js$$' | xargs cat | wc -l
