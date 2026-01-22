package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/wolfmanjm/svd_web/assets"
	"github.com/wolfmanjm/svd_web/gen/dbstore"
)

var useBrowser bool = false

func main() {

	if useBrowser {
		mux := http.NewServeMux()

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			_ = assets.SiteLayout().Render(r.Context(), w)
		})

		fmt.Println("Server starting on port 8080...")
		if err := http.ListenAndServe(":8080", mux); err != nil {
			log.Fatal(err)
		}
	} else {
		periphs := []string {"Uart0", "Uart1", "SPI0", "SPI1"}
		assets.TestIteration(periphs).Render(context.Background(), os.Stdout)
		//assets.SiteLayout().Render(context.Background(), os.Stdout)
	}

	err := run()
	if err != nil {
		panic(err)
	}
}

func run() error {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, "host=localhost port=5432 user=morris password=test dbname=svd sslmode=disable")
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	queries := dbstore.New(conn)

	// list all mpus
	mpus, err := queries.ListMPUs(ctx)
	if err != nil {
		return err
	}
	log.Println(mpus)

	return nil
}
