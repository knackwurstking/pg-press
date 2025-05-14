all: init build

include .env

BINARY_NAME := "picow-led"
SERVER_APP_NAME := ${BINARY_NAME}

clean:
	git clean -xfd

init:
	npm install
	cp ./node_modules/ui/dist/ui.css ./public/css/ui-v4.2.0.css
	mkdir -p ./public/js
	cp ./node_modules/ui/dist/ui.min.umd.cjs ./public/js/ui-v4.2.0.min.umd.cjs
	go mod tidy -v

generate:
	# NOTE: Install eslint with `npm init @eslint/config@latest`
	go mod tidy -v
	rm -rf ./public/js/assets/*
	npx vite build --config ./vite.config.js
	npx vite build --config ./vite.pwa.config.js

dev:
	make generate
	which gow || (echo 'gow is not installed, install with: `go install github.com/mitranim/gow@latest`' && exit 1)
	gow -e=go,html,js,json -v -r run . server -a :8887

run:
	make generate
	go run . server -a :8887

test:
	go test -v ./...

build:
	make test
	make generate
	go build -v -ldflags="-w -s" -o bin/${BINARY_NAME}

generate-pwa-assets:
	npx pwa-assets-generator

# NOTE: Standard systemd stuff

define SYSTEMD_SERVICE_FILE
[Unit]
Description=Control my fucking lights
After=network.target

[Service]
EnvironmentFile=%h/.config/picow-led/.env
ExecStart=${SERVER_APP_NAME} server

[Install]
WantedBy=default.target
endef

UNAME := $(shell uname)
check-linux:
ifneq ($(UNAME), Linux)
	@echo 'This won’t work here since you’re not on Linux.'
	@exit 1
endif

export SYSTEMD_SERVICE_FILE
install: check-linux
	echo "$$SYSTEMD_SERVICE_FILE" > ${HOME}/.config/systemd/user/${SERVER_APP_NAME}.service
	systemctl --user daemon-reload
	echo "--> Created a service file @ ${HOME}/.config/systemd/user/${SERVER_APP_NAME}.service"
	sudo cp ./bin/${SERVER_APP_NAME} /usr/local/bin/

start: check-linux
	systemctl --user restart ${SERVER_APP_NAME}

stop: check-linux
	systemctl --user stop ${SERVER_APP_NAME}

log: check-linux
	journalctl --user -u ${SERVER_APP_NAME} --follow --output cat
