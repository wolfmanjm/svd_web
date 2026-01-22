.PHONY: genassets
genassets:
	goht generate --path assets

.PHONY: migrate
migrate:
	goose postgres "host=localhost port=5432 user=morris password=test dbname=svd sslmode=disable" -dir database up

.PHONY: generate
generate:
	sqlc generate
