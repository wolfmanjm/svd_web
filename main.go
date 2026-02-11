package main

import (
	"embed"
	"flag"
	"fmt"
	"os"

	"github.com/wolfmanjm/svd_web/cmd/parse-svd"
	"github.com/wolfmanjm/svd_web/cmd/svd_server"
)

//go:embed files/*
var staticFiles embed.FS

func main() {
	// you can override the connection string for pgsql with this env
	url, ok := os.LookupEnv("PSQLURL")
	if !ok {
		url = "host=pi5.local port=5432 user=morris password=test dbname=svd sslmode=disable"
	}

	// get flags
	dburl := flag.String("dburl", url, "Set DB URL")
	port := flag.Int("port", 8080, "Set server listen port")
	flag.Parse()

	cmd := flag.Arg(0)
	if cmd == "" {
		fmt.Println("Usage: main [add-svd|serve]")
		os.Exit(0)
	}

	if cmd == "add-svd" {
		svnPath := flag.Arg(1)
		if svnPath == "" {
			fmt.Println("Usage: main add-svd svn-path")
			os.Exit(1)
		}
		err := parse_svd.Convert(svnPath, *dburl)
		if err != nil {
			panic(err)
		}
		os.Exit(0)
	}

	// default is run server
	err := svd_server.Server(*dburl, staticFiles, *port)
	if err != nil {
		panic(err)
	}
}
