package svd_server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"

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

	// example of accessing all the peripherals for a MCU
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

	// example of accessing all the registers for a peripheral of a mcu
	mpu := "rp2350"
	m, err := queries.FindMCU(ctx, mpu)
	if err != nil {
		return fmt.Errorf("find MCU %s - %w", mpu, err)
	}

	periph := "uart0"
	fmt.Printf("\nRegisters for peripheral %s of %s\n", periph, m.Name)

	f := dbstore.FindPeripheralParams {
		MpuID: m.ID,
		Name: periph,
	}
	p, err := queries.FindPeripheral(ctx, f)
	if err != nil {
		return fmt.Errorf("find Peripheral %s - %w", periph, err)
	}

	r, err := queries.FetchRegisters(ctx, p.ID)
	if err != nil {
		return fmt.Errorf("fetch Registers for peripheral %s - %w", periph, err)
	}
	var names []string
	for _, x := range r {
		names = append(names, x.Name)
	}
	fmt.Println(strings.Join(names, ", "))

	// Example of accessing all the fields for a register
	reg := "uartcr"
	fmt.Printf("\nFields for register %s, peripheral %s of %s\n", reg, periph, m.Name)

	fr := dbstore.FindRegisterParams {
		PeripheralID: p.ID,
		Name: reg,
	}
	r1, err := queries.FindRegister(ctx, fr)
	if err != nil {
		return fmt.Errorf("find Register %s - %w", reg, err)
	}

	fields, err := queries.FetchFields(ctx, r1.ID)
	if err != nil {
		return fmt.Errorf("fetch Feilds for register %s - %w", r1.Name, err)
	}
	w := tabwriter.NewWriter(os.Stdout, 10, 4, 2, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "Name\tNumBits\tBitOffset\tDescription")
	for _, f := range fields {
		fmt.Fprintf(w, "%s\t%d\t%d\t%s\n", f.Name, f.NumBits, f.BitOffset, f.Description.String[0:80])
	}
	w.Flush()

	return nil
}
