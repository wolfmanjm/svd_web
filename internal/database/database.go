package database

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/jackc/pgx/v5"
	"github.com/wolfmanjm/svd_web/gen/dbstore"
)

type Database struct {
	queries *dbstore.Queries
	conn    *pgx.Conn
	ctx     context.Context
}

func Setup(cstr string) (*Database, error) {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, cstr)
	if err != nil {
		return nil, err
	}

	var db Database
	db.ctx = ctx
	db.conn = conn
	db.queries = dbstore.New(conn)

	return &db, nil
}

func (db *Database) Close() {
	db.conn.Close(db.ctx)
}

// yet another layer of abstraction
func (db *Database) GetMpus() ([]dbstore.Mpu, error) {
	// get all mpus
	mpus, err := db.queries.ListMPUs(db.ctx)
	if err != nil {
		return mpus, fmt.Errorf("getMpus - %w", err)
	}
	return mpus, nil
}

// return the single MCU with the given ID
func (db *Database) GetMpu(id int32) dbstore.Mpu {
	m, _ := db.queries.GetMCU(db.ctx, id)
	return m
}

// return the peripherals for the given MCUID
func (db *Database) GetPeripherals(id int32) ([]dbstore.Peripheral, error) {
	p, err := db.queries.FetchPeripherals(db.ctx, id)
	if err != nil {
		return p, fmt.Errorf("getPeripherals - %w", err)
	}
	return p, nil
}

// return the single peripheral given the ID
func (db *Database) GetPeripheral(pid int32) dbstore.Peripheral {
	p, _ := db.queries.GetPeripheral(db.ctx, pid)
	return p
}

// return peripherals that match the pattern
func (db *Database) FindPeripherals(id int32, pat string) ([]dbstore.Peripheral, error) {
	args := dbstore.FindPeripheralsParams {
		MpuID: id,
		Name: "%" + pat + "%",
	}
	p, err := db.queries.FindPeripherals(db.ctx, args)
	if err != nil {
		return p, fmt.Errorf("findPeripherals - %w", err)
	}
	return p, nil
}

// Get registers for the specified peripheral, will follow the DerivedFrom field if needed
func (db *Database) GetRegisters(pid int32) ([]dbstore.Register, error) {
	// if this peripheral is derived from, then we need to get the registers from that peripheral
	p := db.GetPeripheral(pid)
	if p.DerivedFromID.Valid {
		pid = p.DerivedFromID.Int32
	}

	r, err := db.queries.FetchRegisters(db.ctx, pid)
	if err != nil {
		return r, fmt.Errorf("fetch Registers - %w", err)
	}
	return r, nil
}

// return the single register given the ID
func (db *Database) GetRegister(rid int32) dbstore.Register {
	r, _ := db.queries.GetRegister(db.ctx, rid)
	return r
}

// Get fields for the specified register
func (db *Database) GetFields(rid int32) ([]dbstore.Field, error) {

	fields, err := db.queries.FetchFields(db.ctx, rid)
	if err != nil {
		return fields, fmt.Errorf("fetch Fields - %w", err)
	}
	return fields, nil
}

func (db *Database) DoStuff() error {
	// list all mpus
	mpus, err := db.queries.ListMPUs(db.ctx)
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
		p, err := db.queries.FetchPeripherals(db.ctx, m.ID)
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
	m, err := db.queries.FindMCU(db.ctx, mpu)
	if err != nil {
		return fmt.Errorf("find MCU %s - %w", mpu, err)
	}

	periph := "uart0"
	fmt.Printf("\nRegisters for peripheral %s of %s\n", periph, m.Name)

	f := dbstore.FindPeripheralParams{
		MpuID: m.ID,
		Name:  periph,
	}
	p, err := db.queries.FindPeripheral(db.ctx, f)
	if err != nil {
		return fmt.Errorf("find Peripheral %s - %w", periph, err)
	}

	r, err := db.queries.FetchRegisters(db.ctx, p.ID)
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

	fr := dbstore.FindRegisterParams{
		PeripheralID: p.ID,
		Name:         reg,
	}
	r1, err := db.queries.FindRegister(db.ctx, fr)
	if err != nil {
		return fmt.Errorf("find Register %s - %w", reg, err)
	}

	fields, err := db.queries.FetchFields(db.ctx, r1.ID)
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
