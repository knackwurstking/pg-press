.PHONY: clean generate init run dev build count macos-install \
	macos-start-service macos-stop-service macos-restart-service \
	macos-print-service macos-watch-service macos-update

all: init build

BINARY_NAME := pg-press

SERVER_ADDR := :9020
SERVER_ADDR_DEV := :8888
SERVER_PATH_PREFIX := /$(BINARY_NAME)

BIN_DIR := ./bin

INSTALL_PATH := /usr/local/bin

APP_DATA := $(HOME)/Library/Application Support/$(BINARY_NAME)
SERVICE_FILE := $(HOME)/Library/LaunchAgents/com.$(BINARY_NAME).plist
LOG_FILE := $(APP_DATA)/pg-press.log

clean:
	git clean -xfd

generate:
	templ generate

init: generate
	go mod tidy -v
	git submodule init
	git submodule update --recursive

run: generate
	SERVER_PATH_PREFIX=${SERVER_PATH_PREFIX} \
		go run ./cmd/${BINARY_NAME} server -a ${SERVER_ADDR}

dev:
	which gow || \
		( \
			echo 'gow is not installed, install with: `go install github.com/mitranim/gow@latest`' && \
			exit 1 \
		)
	export LOG_LEVEL=debug && \
	export LOG_FORMAT=text && \
	export SERVER_PATH_PREFIX=${SERVER_PATH_PREFIX} && \
	gow -e=go,json,html,js,css -r run ./cmd/${BINARY_NAME} server --addr ${SERVER_ADDR_DEV}

build:
	go build -v -o ./bin/${BINARY_NAME} ./cmd/${BINARY_NAME}

count:
	find . | grep -e '\.go$$' -e '\.html$$' -e '\.css$$' -e '\.js$$' -e '\.templ$$' | grep --invert-match '_templ\.go$$' | xargs cat | wc -l

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
		<key>LOG_LEVEL</key>
		<string>info</string>
		<key>LOG_FORMAT</key>
		<string>text</string>
		<key>ADMINS</key>
		<string></string>
	</dict>
</dict>
</plist>
endef

export LAUNCHCTL_PLIST

macos-install:
	@echo "Installing $(BINARY_NAME) for macOS..."
	mkdir -p $(INSTALL_PATH)
	sudo cp $(BIN_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	echo "$$LAUNCHCTL_PLIST" > $(SERVICE_FILE)
	@echo "$(BINARY_NAME) installed successfully"

macos-start-service:
	@echo "Starting $(BINARY_NAME) service..."
	launchctl load -w $(SERVICE_FILE)
	launchctl start com.$(BINARY_NAME)

macos-stop-service:
	@echo "Stopping $(BINARY_NAME) service..."
	launchctl stop com.$(BINARY_NAME) || exit 0
	launchctl unload -w $(SERVICE_FILE)

macos-restart-service:
	@echo "Restarting $(BINARY_NAME) service..."
	make macos-stop-service || exit 0
	make macos-start-service

macos-print-service:
	@echo "$(BINARY_NAME) service information:"
	launchctl print gui/$(shell id -u)/com.$(BINARY_NAME) || echo "Service not loaded or running"

macos-watch-service:
	@echo "$(BINARY_NAME) watch server logs @ \"$(LOG_FILE)\":"
	if [ -f "$(LOG_FILE)" ]; then \
		echo "Watching logs... Press Ctrl+C to stop"; \
		tail -f "$(LOG_FILE)"; \
	else \
		echo "Log file not found. Make sure the service is running or has been started."; \
		echo "Log file path: \"$(LOG_FILE)\""; \
	fi

macos-update: all
	make macos-stop-service
	sudo cp $(BIN_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	make macos-start-service
	make macos-watch-service
