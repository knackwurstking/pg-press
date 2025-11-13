all: init build

BINARY_NAME := pg-press

SERVER_ADDR := :9020
SERVER_PATH_PREFIX := /$(BINARY_NAME)
APP_DATA := $(HOME)/Library/Application\ Support/pg-press

clean:
	git clean -xfd

init:
	go mod tidy -v
	git submodule init
	git submodule update --recursive

generate:
	templ generate

run:
	make generate
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
	gow -e=go,json,html,js,css -r run ./cmd/${BINARY_NAME} server --addr ${SERVER_ADDR}

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
	<string>com.pg-press</string>

	<key>ProgramArguments</key>
	<array>
		<string>/usr/local/bin/pg-press</string>
		<string>server</string>
	</array>

	<key>RunAtLoad</key>
	<true/>

	<key>KeepAlive</key>
	<true/>

	<key>StandardOutPath</key>
	<string>/var/log/pg-press.log</string>

	<key>StandardErrorPath</key>
	<string>/var/log/pg-press.log</string>

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
	mkdir -p /usr/local/bin
	cp ./bin/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	chmod +x /usr/local/bin/$(BINARY_NAME)
	mkdir -p $(APP_DATA)
	@echo "$$LAUNCHCTL_PLIST" > ~/Library/LaunchAgents/com.pg-press.plist
	@echo "$(BINARY_NAME) installed successfully"

macos-start-service:
	@echo "Starting $(BINARY_NAME) service..."
	launchctl load -w ~/Library/LaunchAgents/com.$(BINARY_NAME).plist
	launchctl start com.$(BINARY_NAME)

macos-stop-service:
	@echo "Stopping $(BINARY_NAME) service..."
	launchctl stop com.$(BINARY_NAME)
	launchctl unload -w ~/Library/LaunchAgents/com.$(BINARY_NAME).plist

macos-restart-service:
	@echo "Restarting $(BINARY_NAME) service..."
	make macos-stop-service
	make macos-start-service

macos-print-service:
	@echo "$(BINARY_NAME) service information:"
	@launchctl print gui/$$(id -u)/com.$(BINARY_NAME) || echo "Service not loaded or running"
