package svd_server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/wolfmanjm/svd_web/assets"
	"github.com/wolfmanjm/svd_web/gen/dbstore"
)

var useBrowser bool = false

func Server(cstr string) error {

	if useBrowser {
		mux := http.NewServeMux()

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			_ = assets.SiteLayout().Render(r.Context(), w)
		})

		fmt.Println("Server starting on port 8080...")
		if err := http.ListenAndServe(":8080", mux); err != nil {
			return err
		}
	} else {
		periphs := []string {"Uart0", "Uart1", "SPI0", "SPI1"}
		assets.TestIteration(periphs).Render(context.Background(), os.Stdout)
		//assets.SiteLayout().Render(context.Background(), os.Stdout)
	}

	return run(cstr)
}

func run(cstr string) error {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, cstr)
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

	// example of etting all the MPUs in the database
	fmt.Println("This database supports the following MPUs")
	for _, m := range mpus {
		fmt.Println(m.Name)
	}

	// example of accessing al the peripherals for a MCU
	for _, m := range mpus {
		p, err := queries.FetchPeripherals(ctx, m.ID)
		if err != nil {
			return err
		}
		fmt.Println("Peripherals for ", m.Name)
		var names []string
		for _, x := range p {
			names = append(names, x.Name)
		}
		fmt.Println(strings.Join(names, ", "))
	}

	return nil
}
