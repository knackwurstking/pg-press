all: init build

SERVER_ADDR := :9020
SERVER_PATH_PREFIX := /pg-vis

clean:
	git clean -xfd
	cd ./frontend && git clean -xfd

init:
	go mod tidy -v
	git submodule init
	git submodule update --recursive
	cd frontend && npm ci 

run:
	go run ./cmd/pg-vis

build:
	go build ./cmd/pg-vis
