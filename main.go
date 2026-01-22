package main

import (
	"os"
	"fmt"
	add_svd "github.com/wolfmanjm/svd_web/cmd/add-svd"
	svd_server "github.com/wolfmanjm/svd_web/cmd/svd_server"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: main [add|serve] [dbfn]")
		os.Exit(0)
	}

	// you can override the connection string for pgsql with this env
	url, ok := os.LookupEnv("PSQLURL")
	if !ok {
		url = "host=pi5.local port=5432 user=morris password=test dbname=svd sslmode=disable"
	}

	cmd := os.Args[1]
	if cmd == "add" {
		if len(os.Args) < 3 {
			fmt.Println("Usage: main add dbfn")
			os.Exit(0)
		}
		err := add_svd.AddSVD(url, os.Args[2])
		if err != nil {
			panic(err)
		}
		os.Exit(1)
	}

	err := svd_server.Server(url)
	if err != nil {
		panic(err)
	}
}
