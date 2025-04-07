clean:
	git clean -fxd

run:
	go mod tidy
	go run ./cmd/pg-vis-server

build:
	go mod tidy
	go build -o ./bin/pg-vis-server ./cmd/pg-vis-server
