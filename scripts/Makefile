.PHONY: convert_to_json, test_convert_to_json, convert_to_sql, test_convert_to_sql

convert_to_json:
	@go run -v ./convert_to_json -clean ~/Library/Application\ Support/pg-press/pg-press.db

test_convert_to_json:
	@go run -v ./convert_to_json -clean ./test/pg-press.db


convert_to_sql:
	@go run -v ./convert_to_sql -clean

test_convert_to_sql:
	@go run -v ./convert_to_sql -clean
