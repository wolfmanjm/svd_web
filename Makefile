DSN?="host=pi5.local port=5432 user=morris password=test dbname=svd sslmode=disable"

.PHONY: assets
assets:			## build all the HAML assets
	goht generate --path assets

migrate/up:		## migrate up
	goose postgres ${DSN} -dir data/sql/migrations up

migrate/down:		## Migrate down to previous migration
	goose postgres ${DSN} -dir data/sql/migrations down

migrate/fresh: 		## Create a fresh database
	@echo "Dropping..."
	goose postgres ${DSN} -dir data/sql/migrations reset
	@make migrate/up

sql:			## Generate the sqlc helpers
	sqlc generate

build-rpi:		## build for rpi
	OOS=linux GOARCH=arm64 go build -o svd_web_rpi

build:			## build
	go build -o svd_web

add_svds:		## Add all the known SVDs to the database

add:			## Add a SVD to the database name=path-to-svd
	@if [ -z "${name}" ] ; then echo "Need name"; exit 1; fi;
	go run main.go add-svd ${name}

serve: assets		## Run the server, builds assets first
	go run main.go serve

check:              	## do a static check on the entore project
	staticcheck ./...

help:           	## Show this help.
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'
