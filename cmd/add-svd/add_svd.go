package send_svd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/wolfmanjm/svd_web/gen/dbstore"
)

// This transfers the data in the local svd database to the postgresql database
func AddSVD(url, fn string) error {
	err := run(url, fn)
	if err != nil {
		return err
	}
	return nil
}

func run(url, fn string) error {
	// open the local database
	ldb, err := OpenLocalDatabase(fn)
	if err != nil {
		return err
	}

	mpus, err := fetch_mpus(ldb)
	if err != nil {
		return err
	}

	mpu_id := mpus[0].id

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	queries := dbstore.New(conn)

	// now populate the psql database
	fmt.Println("Populating MPU: ", mpus[0].name)
	m := dbstore.AddMCUParams{
		Name: mpus[0].name,
		Description: pgtype.Text {String: mpus[0].description.V, Valid: mpus[0].description.Valid},
	}

	r_mpuid, err := queries.AddMCU(ctx, m)
	if err != nil {
		return fmt.Errorf("Failed to add MCU: %w", err)
	}
	periphs, err := fetch_peripherals(ldb, mpu_id)
	if err != nil {
		return fmt.Errorf("failed to fetch peripherals - %w", err)
	}

	for _, p := range periphs {
		fmt.Println("  Populating Peripheral: ", p.name)

		r_p := dbstore.AddPeripheralParams {
			MpuID: r_mpuid,
			Name: p.name,
			DerivedFromID: pgtype.Int4{ Int32: int32(p.derived_from.V), Valid: p.derived_from.Valid, },
			BaseAddress: p.base_address,
			Description: pgtype.Text { String: p.description.V, Valid: p.description.Valid },
		}

		r_pid, err := queries.AddPeripheral(ctx, r_p)
		if err != nil {
			return fmt.Errorf("Failed to add Peripheral for mpu_id %d - %w", r_mpuid, err)
		}

		// This collects all the registers and their fields
		pr, err := collect_registers(ldb, mpu_id, p.id)
		if err != nil {
			return fmt.Errorf("Error in collect registers - %w", err)
		}

		// this will populate the registers and their fields
		err = populate_registers(ctx, queries, r_pid, pr.registers)
		if err != nil {
			return fmt.Errorf("Error in populate_registers - %w", err)
		}
	}

	return nil
}

func populate_registers(ctx context.Context, queries *dbstore.Queries, pid int32, registers *[]Register) error {
	if registers == nil { return nil }
	for _, r := range *registers  {
		// fmt.Printf("    Populating Register %s for peripheral_id %d\n", r.name, pid)

 		rparams := dbstore.AddRegisterParams {
 			PeripheralID: pid,
 			Name: r.name,
 			AddressOffset: r.address_offset,
 			ResetValue: pgtype.Text { String: r.reset_value.V, Valid: r.reset_value.Valid },
 			Description: pgtype.Text { String: r.description.V, Valid: r.description.Valid },
 		}

		r_rid, err := queries.AddRegister(ctx, rparams)
		if err != nil {
			return fmt.Errorf("Failed to Add Register %s for peripheral_id %d - %w", r.name, pid, err)
		}

		err = populate_fields(ctx, queries, r_rid, r.fields)
		if err != nil {
			return fmt.Errorf("Error in populate_fields - %w", err)
		}
	}

	return nil
}

func populate_fields(ctx context.Context, queries *dbstore.Queries, rid int32, fields *[]Field) error {
	if fields == nil { return nil }
	for _, f := range *fields {
		// fmt.Printf("      Populating Field %s for register_id %d\n", f.name, rid)
		fparams := dbstore.AddFieldParams {
			RegisterID: rid,
			Name: f.name,
			NumBits: int32(f.num_bits),
			BitOffset: int32(f.bit_offset),
 			Description: pgtype.Text { String: f.description.V, Valid: f.description.Valid },
		}
		_, err := queries.AddField(ctx, fparams)
		if err != nil {
			return fmt.Errorf("Failed to Add Field %s for register_id %d - %w", f.name, rid, err)
		}
	}

	return nil
}

// This is all done by hand as it predates my use of sqlc
type BasicInfo struct {
	id int
	name string
	description sql.Null[string]
}

type MPU struct {
	BasicInfo
}

type Peripheral struct {
	BasicInfo
	derived_from sql.Null[int]
	base_address string
	registers *[]Register
}

type Register struct {
	BasicInfo
	address_offset string
	reset_value sql.Null[string]
	fields *[]Field
}

type Field struct {
	BasicInfo
	num_bits int
	bit_offset int
}

// print helpers for the structs
func (f Field) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Field:  %v, number bits %v, bit offset: %v\n", f.name, f.num_bits, f.bit_offset)
	return b.String()
}

func (r Register) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Register: %v, address offset: %v\n", r.name, r.address_offset)
	if r.fields != nil {
		for _, f := range *r.fields  {
			fmt.Fprint(&b, "    ", f)
		}
	}

	return b.String()
}

func (p Peripheral) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Peripheral: %v, base address: %v\n", p.name, p.base_address)
	if p.registers != nil {
		for _, r := range *p.registers  {
			fmt.Fprint(&b, "  ", r)
		}
	}
	return b.String()
}

func OpenLocalDatabase(dbfn string) (*sql.DB, error) {
	// make sure database file exists
	_, err := os.Stat(dbfn)
	if err != nil {
		return nil, fmt.Errorf("database file %v does not exist - %w", dbfn, err)
	}

	fn := "file:" + dbfn + "?mode=ro"
	db, err := sql.Open("sqlite", fn)
	if err != nil {
		return nil, fmt.Errorf("Unable to open database file %v - %w", dbfn, err)
	}
	// makes sure the database is ok
	mpus, err := fetch_mpus(db)
	if err != nil {
		return nil, fmt.Errorf("Unable to get MPUs, is the database valid? - %w", err)
	}

	if len(mpus) < 1 {
		return nil, errors.New("No MPUs in database")
	}

	return db, nil
}

func CloseLocalDatabase(db *sql.DB) {
	db.Close()
}

func IntPow(base, exp int) int {
    result := 1
    for {
        if exp & 1 == 1 {
            result *= base
        }
        exp >>= 1
        if exp == 0 {
            break
        }
        base *= base
    }

    return result
}

// collect all the registers and their fields for the named peripheral
func collect_registers(db *sql.DB, mpu_id, periph_id int) (Peripheral, error) {
	p, err := fetch_peripheral(db, mpu_id, periph_id)
	if err != nil {
		return p, fmt.Errorf("Peripheral %v not found: %w", periph_id, err)
	}

	id := p.id

	if p.derived_from.Valid {
		id = p.derived_from.V
	}

	regs, err := fetch_registers(db, id)
	if err != nil {
		return p, err
	}
	for i, r := range regs {
		fields, err := fetch_fields(db, r.id)
		if err != nil {
			return p, err
		}
		regs[i].fields = &fields
	}

	p.registers = &regs

	return p, nil
}

// local database helpers
func fetch_mpus(db *sql.DB) ([]MPU, error) {
	var mpus []MPU
	mpu_rows, err := db.Query("select id, name, description from mpus ORDER BY name")
	if err != nil {
		return mpus, err
	}
	defer mpu_rows.Close()

	for mpu_rows.Next() {
		var m MPU
		if err := mpu_rows.Scan(&m.id, &m.name, &m.description); err != nil {
			return mpus, err
		}
		mpus= append(mpus, m)
	}

	if err := mpu_rows.Err(); err != nil {
		return mpus, err
	}

	return mpus, nil
}

func fetch_peripherals(db *sql.DB, mpu_id int) ([]Peripheral, error) {
	var periphs []Peripheral
	periph_rows, err := db.Query("select id, derived_from_id, name, base_address, description from peripherals WHERE mpu_id = ? ORDER BY name", mpu_id)
	if err != nil {
		return periphs, err
	}
	defer periph_rows.Close()

	for periph_rows.Next() {
		var p Peripheral
		err = periph_rows.Scan(&p.id, &p.derived_from, &p.name, &p.base_address, &p.description)
		if err != nil {
			return periphs, err
		}
		periphs= append(periphs, p)
	}

	if err := periph_rows.Err(); err != nil {
		return periphs, err
	}

	return periphs, nil
}

func fetch_peripheral(db *sql.DB, mpu_id, id int) (Peripheral, error) {
	var p Peripheral

    if err := db.QueryRow("SELECT id, derived_from_id, name, base_address, description from peripherals WHERE mpu_id = ? AND id = ?", mpu_id, id).
    	Scan(&p.id, &p.derived_from, &p.name, &p.base_address, &p.description); err != nil {
        	return p, err
    }
    return p, nil;
}

func fetch_registers(db *sql.DB, p_id int) ([]Register, error) {
	register_rows, err := db.Query("select id, name, address_offset, reset_value, description from registers WHERE peripheral_id = ? ORDER BY name", p_id)

	if err != nil {
		return nil, fmt.Errorf("failure in fetch_registers query for id %v: %w", p_id, err)
	}
	defer register_rows.Close()

	var registers []Register
	for register_rows.Next() {
		var reg Register
		err = register_rows.Scan(&reg.id, &reg.name, &reg.address_offset, &reg.reset_value, &reg.description)
		if err != nil {
			return nil, fmt.Errorf("failure in fetch_registers scan for id %v: %w", p_id, err)
		}
		registers= append(registers, reg)
	}

	if err := register_rows.Err(); err != nil {
		return nil, fmt.Errorf("failure in fetch_registers rows for id %v: %w", p_id, err)
	}

	return registers, nil
}

func fetch_fields(db *sql.DB, r_id int) ([]Field, error) {
	field_rows, err := db.Query("select name, num_bits, bit_offset, description from fields WHERE register_id = ? ORDER BY bit_offset", r_id)

	if err != nil {
		return nil, fmt.Errorf("failure in fetch_fields query for id %v: %w", r_id, err)
	}
	defer field_rows.Close()
	var fields []Field
	for field_rows.Next() {
		var f Field
		err = field_rows.Scan(&f.name, &f.num_bits, &f.bit_offset, &f.description)
		if err != nil {
			return nil, fmt.Errorf("failure in fetch_fields scan for id %v: %w", r_id, err)
		}
		fields= append(fields, f)
	}

	if err := field_rows.Err(); err != nil {
		return nil, fmt.Errorf("failure in fetch_registers rows for id %v: %w", r_id, err)
	}

	return fields, nil
}
