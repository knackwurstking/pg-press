clean:
	git clean -fxd

run:
	go mod tidy || exit $?
	go run ./cmd/pg-vis-server

build:
	go mod tidy || exit $?
	go build -o ./bin/pg-vis-server ./cmd/pg-vis-server

install:
	make build || exit $?
	cp ./cmd/pg-vis-server/pg-vis-server.service ${HOME}/.config/systemd/user/ || exit $?
	systemctl --user daemon-reload || exit $?
	@echo 'Start the service with `systemctl --user start pg-vis-server`'
	@echo 'Check the logs with `journal --user -u pg-vis-server --follow --output cat`'
