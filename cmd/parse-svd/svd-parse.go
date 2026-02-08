package parse_svd

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/wolfmanjm/svd_web/gen/dbstore"
)

// SVD structure definitions based on CMSIS-SVD specification
type Device struct {
	XMLName     xml.Name     `xml:"device"`
	Name        string       `xml:"name"`
	Description string       `xml:"description"`
	Peripherals []Peripheral `xml:"peripherals>peripheral"`
}

type Peripheral struct {
	Name         string       `xml:"name"`
	Description  string       `xml:"description"`
	BaseAddress  string       `xml:"baseAddress"`
	GroupName    string       `xml:"groupName"`
	Registers    []Register   `xml:"registers>register"`
	DerivedFrom  string       `xml:"derivedFrom,attr"`
	AddressBlock AddressBlock `xml:"addressBlock"`
}

type AddressBlock struct {
	Offset string `xml:"offset"`
	Size   string `xml:"size"`
	Usage  string `xml:"usage"`
}

type Register struct {
	Name        string  `xml:"name"`
	Description string  `xml:"description"`
	Offset      string  `xml:"addressOffset"`
	Size        string  `xml:"size"`
	Access      string  `xml:"access"`
	ResetValue  string  `xml:"resetValue"`
	Fields      []Field `xml:"fields>field"`
}

type Field struct {
	Name        string `xml:"name"`
	Description string `xml:"description"`
	BitOffset   string `xml:"bitOffset"`
	BitWidth    string `xml:"bitWidth"`
	BitRange    string `xml:"bitRange"`
	Access      string `xml:"access"`
	LSB         string `xml:"lsb"`
	MSB         string `xml:"msb"`
	Enumerations []Enumeration `xml:"enumeratedValues>enumeratedValue"`
}

type Enumeration struct {
	Name 		string `xml:"name"`
    Value		string `xml:"value"`
	Description string `xml:"description"`
}

// This reads in a SVD XML file and UnMarshals it into the structures
func Convert(filename, url string) error {
	// Read the SVD file
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("in convert reading file: %w", err)
	}

	// Parse XML
	var device Device
	err = xml.Unmarshal(data, &device)
	if err != nil {
		return fmt.Errorf("in convert parsing XML: %w", err)
	}

	if err = WriteToDatabase(device, url); err != nil {
		return fmt.Errorf("in Writing to Database: %w", err)
	}

	return nil
}

// keeps a list of peripherals to id mapping, needed for derived_from peripherals
var periph_ids map[string]int32

func WriteToDatabase(device Device, url string) error {
	// open the database connection
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	// get queries
	queries := dbstore.New(conn)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// setup the transaction
	qtx := queries.WithTx(tx)

	// now populate the psql database
	fmt.Println("Populating Device: ", device.Name)
	m := dbstore.AddMCUParams{
		Name: device.Name,
		Description: pgtype.Text {String: device.Description, Valid: device.Description != ""},
	}

	mpuid, err := qtx.AddMCU(ctx, m)
	if err != nil {
		return fmt.Errorf("failed to add MCU: %w", err)
	}

	periph_ids = make(map[string]int32)

	if err := addPeripherals(ctx, qtx, mpuid, device.Peripherals); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func addPeripherals(ctx context.Context, qtx *dbstore.Queries, mpuid int32, peripherals []Peripheral) error {
	// insert peripherals and their registers, fields and enumerations
	for i := range peripherals {
		p := peripherals[i]
		fmt.Println("Populating Peripheral: ", p.Name)

		var isDerivedFrom bool
		var d_id int32
		if p.DerivedFrom != "" {
			d_id, isDerivedFrom = periph_ids[p.DerivedFrom]
			if !isDerivedFrom {
				return fmt.Errorf("derived from %v is not in ids", p.DerivedFrom)
			}
		} else {
			isDerivedFrom = false
			d_id = 0
		}

		pp := dbstore.AddPeripheralParams {
			MpuID: mpuid,
			Name: p.Name,
			BaseAddress: p.BaseAddress,
			DerivedFromID: pgtype.Int4{ Int32: d_id, Valid: isDerivedFrom, },
			Description: pgtype.Text { String: p.Description, Valid: p.Description != ""},
		}

		pid, err := qtx.AddPeripheral(ctx, pp)
		if err != nil {
			return fmt.Errorf("addPeripherald for mpu_id %d - %w", mpuid, err)
		}

		// record the id of every poeripheral in case it is derived from
		periph_ids[p.Name] = pid

		if !isDerivedFrom {
			// this will populate the registers and their fields
			err = addRegisters(ctx, qtx, pid, p.Registers)
			if err != nil {
				return fmt.Errorf("addRegisters - %w", err)
			}
		}
	}

	return nil
}

func addRegisters(ctx context.Context, qtx *dbstore.Queries, pid int32, registers []Register) error {
	if registers == nil { return nil }

	for i := range registers  {
		r := registers[i]
		// fmt.Printf("  Populating Register %s for peripheral_id %d\n", r.name, pid)

 		rparams := dbstore.AddRegisterParams {
 			PeripheralID: pid,
 			Name: r.Name,
 			AddressOffset: r.Offset,
 			ResetValue: pgtype.Text { String: r.ResetValue, Valid: r.ResetValue != "" },
 			Access: pgtype.Text { String: r.Access, Valid: r.Access != "" },
 			Description: pgtype.Text { String: r.Description, Valid: r.Description != "" },
 		}

		rid, err := qtx.AddRegister(ctx, rparams)
		if err != nil {
			return fmt.Errorf("failed to Add Register %s for peripheral_id %d - %w", r.Name, pid, err)
		}

		err = addFields(ctx, qtx, rid, r.Fields)
		if err != nil {
			return fmt.Errorf("addFields - %w", err)
		}
	}

	return nil
}

func addFields(ctx context.Context, qtx *dbstore.Queries, rid int32, fields []Field) error {
	if fields == nil { return nil }

	for i := range fields {
		f := fields[i]
		// fmt.Printf("    Populating Field %s for register_id %d\n", f.Name, rid)
		bit_offset, num_bits, err := convertBits(f)
		if err != nil {
			return err
		}

		fparams := dbstore.AddFieldParams {
			RegisterID: rid,
			Name: f.Name,
			NumBits: int32(num_bits),
			BitOffset: int32(bit_offset),
 			Access: pgtype.Text { String: f.Access, Valid: f.Access != "" },
 			Description: pgtype.Text { String: f.Description, Valid: f.Description != "" },
		}

		fid, err := qtx.AddField(ctx, fparams)
		if err != nil {
			return fmt.Errorf("failed to Add Field %s for register_id %d - %w", f.Name, rid, err)
		}

		if f.Enumerations != nil {
			err = addEnums(ctx, qtx, fid, f.Enumerations)
			if err != nil {
				return fmt.Errorf("addEnums - %w", err)
			}
		}
	}

	return nil
}

// Handle different bit position formats and convert to num_bits and bit_offset
func convertBits(f Field) (int, int, error) {
	var bit_offset, num_bits int
	var err error

	if f.BitOffset != "" && f.BitWidth != "" {
		bit_offset, err = strconv.Atoi(f.BitOffset)
		if err != nil {
			return 0, 0, fmt.Errorf("in convertBits converting bitOffset %v to integer: %w", f.BitOffset, err)
		}
		num_bits, err = strconv.Atoi(f.BitWidth)
		if err != nil {
			return 0, 0, fmt.Errorf("in convertBits converting bitWidth %v to integer: %w", f.BitWidth, err)
		}

	} else if f.BitRange != "" {
		// Bit Range: [31:16]
		br := strings.Split(f.BitRange[1:len(f.BitRange)-1], ":")
		hr, err := strconv.Atoi(br[0])
		if err != nil {
			return 0, 0, fmt.Errorf("in convertBits converting bitRange to integer: %w",  err)
		}
		lr, err := strconv.Atoi(br[1])
		if err != nil {
			return 0, 0, fmt.Errorf("in convertBits converting bitRange to integer: %w",  err)
		}
		bit_offset = lr
		num_bits = (hr-lr)+1

	} else if f.LSB != "" && f.MSB != "" {
		return 0, 0, fmt.Errorf("in convertBits bit MSB/LSB not handled")

	} else {
		return 0, 0, fmt.Errorf("in convertBits no valid bit info found")
	}

	return bit_offset, num_bits, nil
}


func addEnums(ctx context.Context, qtx *dbstore.Queries, fid int32, enumerations []Enumeration) error {
	for i := range enumerations {
		e := enumerations[i]

		params := dbstore.AddEnumerationParams {
			FieldID: fid,
			Name: e.Name,
			Value: e.Value,
 			Description: pgtype.Text { String: e.Description, Valid: e.Description != "" },
		}

		_, err := qtx.AddEnumeration(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to Add Enum %s for field_id %d - %w", e.Name, fid, err)
		}
	}

	return nil
}

