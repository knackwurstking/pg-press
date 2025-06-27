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
	SERVER_PATH_PREFIX=${SERVER_PATH_PREFIX} \
		go run ./cmd/pg-vis server -a ${SERVER_ADDR}

build:
	go build -v -o ./bin/pg-vis ./cmd/pg-vis
