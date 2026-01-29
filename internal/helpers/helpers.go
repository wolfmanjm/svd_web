package helpers

import (
	"fmt"
	"slices"
	"cmp"
	"github.com/wolfmanjm/svd_web/gen/dbstore"
)

type BitField struct {
	Colspan int32
	Value string
	Reserved bool
}

func GenerateFieldDiagram(fields []dbstore.Field) ([]BitField, []BitField) {
	// convert into a map
	fmap := make(map[int32]dbstore.Field, len(fields))
	for _, f := range fields {
		fmap[f.BitOffset] = f
	}

	// find the maximum bit offset
	maxBit := slices.MaxFunc(fields, func(a, b dbstore.Field) int {
		return cmp.Compare(a.BitOffset, b.BitOffset)
	})

	maxBits := maxBit.BitOffset + maxBit.NumBits-1
	colspan := 31 - maxBits

	var colspans [32]int32
	// Unused bits at MSB of word
	if colspan > 1 {
		colspans[maxBits+1] = colspan
	}

	// find the colspan for each other bit
	for i := int32(0); i <= maxBits; i++ {
		f, ok := fmap[i]
		if ok {
			colspans[i] = f.NumBits
		} else {
			colspans[i] = 1
		}
	}

	// Generate Bit numbers
	var bitNumbers []BitField

	for i := int32(0); i < 32; {
		var b BitField
		b.Colspan = colspans[i]
		if colspans[i] > 1 {
			b.Value = fmt.Sprintf("%d-%d", i+colspans[i]-1, i)
			i += colspans[i]
		} else {
			b.Value = fmt.Sprintf("%d", i)
			i++
		}
		bitNumbers = append(bitNumbers, b)
	}

	// Generate Field names
 	var bitNames []BitField
	for i := int32(0); i < 32; {
		var b BitField
		f, ok := fmap[i]
		b.Colspan = colspans[i]
		b.Value = f.Name
		b.Reserved = !ok
		if colspans[i] > 1 {
			if !ok {
				b.Value = "Reserved"
			}
			i += colspans[i]

		} else {
			if !ok {
				b.Value = "-"
			}
			i++
		}
		bitNames = append(bitNames, b)
	}
	slices.Reverse(bitNumbers)
	slices.Reverse(bitNames)
	return bitNumbers, bitNames
}

