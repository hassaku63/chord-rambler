package material_test

import (
	"github.com/hassaku63/chord-rambler/backend/internal/material"
)

// Alterations, for the tables.
const (
	flat2  = -2
	flat   = -1
	nat    = 0
	sharp  = 1
	sharp2 = 2
)

// note builds a note name for a test table. The tables hold only valid
// values, so a failure here is a mistake in the table.
func note(l material.Letter, alter int) material.NoteName {
	n, err := material.NewNoteName(l, alter)
	if err != nil {
		panic(err)
	}
	return n
}

// interval builds an interval for a test table, like note.
func interval(number int, q material.Quality) material.Interval {
	iv, err := material.NewInterval(number, q)
	if err != nil {
		panic(err)
	}
	return iv
}

// allNotes returns the 35 note names: seven letters, double flat to double
// sharp.
func allNotes() []material.NoteName {
	var notes []material.NoteName
	for l := material.C; l <= material.B; l++ {
		for a := flat2; a <= sharp2; a++ {
			notes = append(notes, note(l, a))
		}
	}
	return notes
}

// allIntervals returns every interval NewInterval accepts.
func allIntervals() []material.Interval {
	qualities := []material.Quality{
		material.Perfect, material.Major, material.Minor, material.Augmented, material.Diminished,
	}
	var ivs []material.Interval
	for number := 1; number <= 13; number++ {
		for _, q := range qualities {
			if iv, err := material.NewInterval(number, q); err == nil {
				ivs = append(ivs, iv)
			}
		}
	}
	return ivs
}
