.PHONY: convert_to_json, convert_to_json_test

convert_to_json:
	@go run -v ./convert_to_json ~/Library/Application\ Support/pg-press/pg-press.db

convert_to_json_test:
	@go run -v ./convert_to_json ./test/pg-press.db
