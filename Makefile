clean:
	git clean -fxd

run:
	go mod tidy
	go run ./cmd/pg-vis
