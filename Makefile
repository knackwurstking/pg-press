clean:
	git clean -fxd

run:
	go mod tidy 
	go run ./cmd/pg-vis-server

build:
	go mod tidy 
	go build -o ./bin/pg-vis-server ./cmd/pg-vis-server

install:
	make build 
	cp ./cmd/pg-vis-server/pg-vis-server.service ${HOME}/.config/systemd/user/ 
	systemctl --user daemon-reload 
	sudo cp ./bin/pg-vis-server /usr/local/bin/pg-vis-server
	@echo 'Start the service with `systemctl --user start pg-vis-server`'
	@echo 'Check the logs with `journalctl --user -u pg-vis-server --follow --output cat`'
