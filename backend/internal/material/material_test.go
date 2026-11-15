package material_test

import (
	"strconv"
	"strings"

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

// notes parses note names for a test table, such as "F♯ A♯ C♯".
func notes(s string) []material.NoteName {
	var out []material.NoteName
	for _, f := range strings.Fields(s) {
		r := []rune(f)
		l := material.Letter(strings.IndexRune("CDEFGAB", r[0]))
		alter := strings.Count(f, "♯") - strings.Count(f, "♭")
		out = append(out, note(l, alter))
	}
	return out
}

// pitchClassesOf returns the pitch-class set of the note names.
func pitchClassesOf(ns []material.NoteName) material.PitchClassSet {
	var s material.PitchClassSet
	for _, n := range ns {
		s = s.Union(material.NewPitchClassSet(n.PitchClass()))
	}
	return s
}

// ivs parses intervals for a test table, such as "M3 P5 m7".
func ivs(s string) []material.Interval {
	qualities := map[byte]material.Quality{
		'P': material.Perfect, 'M': material.Major, 'm': material.Minor,
		'A': material.Augmented, 'd': material.Diminished,
	}
	var out []material.Interval
	for _, f := range strings.Fields(s) {
		q, ok := qualities[f[0]]
		number, err := strconv.Atoi(f[1:])
		if !ok || err != nil {
			panic("bad interval in a test table: " + f)
		}
		out = append(out, interval(number, q))
	}
	return out
}
