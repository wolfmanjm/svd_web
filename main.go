package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/wolfmanjm/svd_web/cmd/svd_server"
	"github.com/wolfmanjm/svd_web/cmd/parse-svd"
)

//go:embed files/*
var staticFiles embed.FS

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: main [add-svd|serve]")
		os.Exit(0)
	}

	// you can override the connection string for pgsql with this env
	url, ok := os.LookupEnv("PSQLURL")
	if !ok {
		url = "host=pi5.local port=5432 user=morris password=test dbname=svd sslmode=disable"
	}

	cmd := os.Args[1]

	if cmd == "add-svd" {
		if len(os.Args) < 3 {
			fmt.Println("Usage: main add-svd svn-path")
			os.Exit(1)
		}
		err := parse_svd.Convert(os.Args[2], url)
		if err != nil {
			panic(err)
		}
		os.Exit(0)
	}

	err := svd_server.Server(url, staticFiles)
	if err != nil {
		panic(err)
	}
}
