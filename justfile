alias b := build
alias s := serve

DSN := "host=pi5.local port=5432 user=morris password=test dbname=svd sslmode=disable"

# migrate up
[group('sql')]
migrate-up:
    goose postgres '{{DSN}}' -dir data/sql/migrations up

# Migrate down to previous migration
[group('sql')]
migrate-down:
    goose postgres '{{DSN}}' -dir data/sql/migrations down

# Create a fresh database
[group('sql')]
migrate-fresh:
    @echo "Dropping..."
    goose postgres '{{DSN}}' -dir data/sql/migrations reset
    @make migrate/up

# Generate the sqlc helpers
[group('sql')]
sql:
    sqlc generate

# build for rpi
[group('build')]
build-rpi:
    OOS=linux GOARCH=arm64 go build -o svd_web_rpi

# build for x86 linux
[group('build')]
build:
    go build -o svd_web

# Add all the known SVDs to the database
add_svds:
    @echo "TBD"

# Add a SVD to the database name=path-to-svd
add name:
    go run main.go add-svd {{name}}

# build all the HAML assets
assets:
    goht generate --path assets

# Run the server, builds assets first
[group('run')]
serve: assets
    go run main.go serve

# do a static check on the entire project
check:
    staticcheck ./...
