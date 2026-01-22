.PHONY: genassets
genassets:
	goht generate --path assets

.PHONY: migrate
migrate:
	goose postgres "host=pi5.local port=5432 user=morris password=test dbname=svd sslmode=disable" -dir data/sql/migrations up
.PHONY: down
down:
	goose postgres "host=pi5.local port=5432 user=morris password=test dbname=svd sslmode=disable" -dir data/sql/migrations down

.PHONY: generate
generate:
	sqlc generate
