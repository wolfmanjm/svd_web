package svd_server

import (
	"context"
	"os"

	"github.com/wolfmanjm/svd_web/assets"
	"github.com/wolfmanjm/svd_web/internal/database"
)


func Test(cstr string) error {

	periphs := []string {"Uart0", "Uart1", "SPI0", "SPI1"}
	assets.TestIteration(periphs).Render(context.Background(), os.Stdout)
	//assets.SiteLayout().Render(context.Background(), os.Stdout)

	// run some database tests
	return run(cstr)
}

func run(cstr string) error {
	db, err := database.Setup(cstr)
	if err != nil {
		return err
	}

	defer db.Close()

	err = db.DoStuff()
	return err
}
