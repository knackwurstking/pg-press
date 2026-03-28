.PHONY: clean generate init run dev build count macos-install macos-update

all: generate init build

BINARY_NAME := pg-press

SERVER_ADDR := :9020
SERVER_ADDR_DEV := :8888
SERVER_PATH_PREFIX := /$(BINARY_NAME)

BIN_DIR := ./bin

INSTALL_PATH := /usr/local/bin

SERVICE_FILE := $(HOME)/Library/LaunchAgents/com.$(BINARY_NAME).plist

APP_DATA := $(HOME)/Library/Application Support/$(BINARY_NAME)
LOG_FILE := $(APP_DATA)/$(BINARY_NAME).log

clean:
	git clean -xfd

generate:
	templ generate
	tailwindcss -i ./internal/assets/public/css/input.css -o ./internal/assets/public/css/output.css

init: generate
	go mod tidy -v
	git submodule init
	git submodule update --recursive

test:
	go test -v ./...

run: generate
	SERVER_PATH_PREFIX=${SERVER_PATH_PREFIX} \
		go run ./cmd/${BINARY_NAME} server -a ${SERVER_ADDR}

dev:
	which gow || \
		( \
			echo 'gow is not installed, install with: `go install github.com/mitranim/gow@latest`' && \
			exit 1 \
		)
	mkdir -p data
	export VERBOSE=true && \
	export SERVER_PATH_PREFIX=${SERVER_PATH_PREFIX} && \
	gow -e=go,json,html,js,css -r run ./cmd/${BINARY_NAME} server --addr ${SERVER_ADDR_DEV} --db data/

build:
	go build -v -o ./bin/${BINARY_NAME} ./cmd/pg-press

count:
	find . -name '*.go' -o -name '*.html' -o -name '*.css' -o -name '*.js' -o -name '*.templ' | grep -v '_templ\.go$$' | xargs cat | wc -l

define LAUNCHCTL_PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.$(BINARY_NAME)</string>

	<key>ProgramArguments</key>
	<array>
		<string>/usr/local/bin/$(BINARY_NAME)</string>
		<string>server</string>
	</array>

	<key>RunAtLoad</key>
	<true/>

	<key>KeepAlive</key>
	<true/>

	<key>StandardOutPath</key>
	<string>$(LOG_FILE)</string>

	<key>StandardErrorPath</key>
	<string>$(LOG_FILE)</string>

	<key>EnvironmentVariables</key>
	<dict>
		<key>SERVER_ADDR</key>
		<string>:9020</string>
		<key>SERVER_PATH_PREFIX</key>
		<string>/pg-press</string>
		<key>ADMINS</key>
		<string></string>
	</dict>
</dict>
</plist>
endef

export LAUNCHCTL_PLIST

macos-install: all
	@echo "Installing $(BINARY_NAME) for macOS..."
	mkdir -p $(INSTALL_PATH)
	sudo cp $(BIN_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	echo "$$LAUNCHCTL_PLIST" > $(SERVICE_FILE)
	@echo "$(BINARY_NAME) installed successfully"

macos-update: all
	sudo cp $(BIN_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Using launchctl command for restarting the service..."
