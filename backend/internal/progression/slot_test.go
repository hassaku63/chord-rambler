package progression_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/progression"
)

const (
	flat  = -1
	nat   = 0
	sharp = 1
)

func note(l material.Letter, alter int) material.NoteName {
	n, err := material.NewNoteName(l, alter)
	if err != nil {
		panic(err)
	}
	return n
}

// ivs parses intervals such as "M3 P5 m7".
func ivs(s string) []material.Interval {
	var out []material.Interval
	for _, f := range strings.Fields(s) {
		q, ok := map[byte]material.Quality{
			'P': material.Perfect, 'M': material.Major, 'm': material.Minor,
			'A': material.Augmented, 'd': material.Diminished,
		}[f[0]]
		number, err := strconv.Atoi(f[1:])
		if !ok || err != nil {
			panic("bad interval in a test table: " + f)
		}
		iv, err := material.NewInterval(number, q)
		if err != nil {
			panic(err)
		}
		out = append(out, iv)
	}
	return out
}

// chord returns the chord on the root with the intervals and, unless empty,
// the bass.
func chord(root material.NoteName, intervals, bass string) material.Chord {
	c, err := material.NewChord(root, ivs(intervals)...)
	if err != nil {
		panic(err)
	}
	if bass != "" {
		if c, err = c.WithBass(ivs(bass)[0]); err != nil {
			panic(err)
		}
	}
	return c
}

func slot(c progression.Content, beats int) progression.Slot {
	s, err := progression.NewSlot(c, beats)
	if err != nil {
		panic(err)
	}
	return s
}

// Only 4/4 can be built (#91), and it has four beats in a bar.
func TestNewMeter(t *testing.T) {
	m, err := progression.NewMeter(4, 4)
	if err != nil || m.BeatsPerBar() != 4 {
		t.Errorf("NewMeter(4, 4) = %v (%d beats), %v; want 4/4 with 4 beats", m, m.BeatsPerBar(), err)
	}
	if m != (progression.Meter{}) {
		t.Errorf("NewMeter(4, 4) = %v; want the zero value", m)
	}
	for _, tt := range [][2]int{{3, 4}, {2, 4}, {6, 8}, {4, 8}, {4, 2}, {0, 4}, {-4, 4}} {
		if _, err := progression.NewMeter(tt[0], tt[1]); !errors.Is(err, progression.ErrMeter) {
			t.Errorf("NewMeter(%d, %d): err = %v; want ErrMeter", tt[0], tt[1], err)
		}
	}
}

// No chord and none chosen yet are different values, and neither holds a
// chord. The zero value is none chosen yet.
func TestContent(t *testing.T) {
	c := chord(note(material.C, nat), "M3 P5", "")
	tests := []struct {
		name                 string
		content              progression.Content
		wantChord            bool
		wantNoChord, wantNot bool
	}{
		{"a chord", progression.ChordOf(c), true, false, false},
		{"no chord", progression.NoChord(), false, true, false},
		{"none chosen yet", progression.Unchosen(), false, false, true},
		{"the zero value", progression.Content{}, false, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := tt.content.Chord()
			if ok != tt.wantChord || (ok && got != c) {
				t.Errorf("Chord() = %v, %v; want a chord: %v", got, ok, tt.wantChord)
			}
			if tt.content.IsNoChord() != tt.wantNoChord || tt.content.IsUnchosen() != tt.wantNot {
				t.Errorf("IsNoChord() = %v, IsUnchosen() = %v; want %v, %v",
					tt.content.IsNoChord(), tt.content.IsUnchosen(), tt.wantNoChord, tt.wantNot)
			}
		})
	}
	if progression.NoChord() == progression.Unchosen() {
		t.Error("NoChord() == Unchosen(); want different values")
	}
	if progression.ChordOf(material.Chord{}) == progression.Unchosen() {
		t.Error("ChordOf(C) == Unchosen(); want different values")
	}
}

// A slot is one beat or longer; the default length is one bar. The zero value
// is none chosen yet, one beat, with no role.
func TestNewSlot(t *testing.T) {
	for _, beats := range []int{1, 3, 4, 8, 16} {
		if s, err := progression.NewSlot(progression.NoChord(), beats); err != nil || s.Beats() != beats {
			t.Errorf("NewSlot(%d) = %d beats, %v; want %d beats", beats, s.Beats(), err, beats)
		}
	}
	for _, beats := range []int{0, -1} {
		if _, err := progression.NewSlot(progression.NoChord(), beats); !errors.Is(err, progression.ErrLength) {
			t.Errorf("NewSlot(%d): err = %v; want ErrLength", beats, err)
		}
	}
	if s := progression.NewBarSlot(progression.Unchosen(), progression.Meter{}); s.Beats() != 4 {
		t.Errorf("NewBarSlot in 4/4 = %d beats; want 4", s.Beats())
	}
	var zero progression.Slot
	if !zero.Content().IsUnchosen() || zero.Beats() != 1 || zero.Role() != progression.RoleNone {
		t.Errorf("zero Slot = %v, %d beats, %v; want none chosen yet, 1 beat, no role",
			zero.Content(), zero.Beats(), zero.Role())
	}
}

// The six structural roles and the four colour roles of the table in
// docs/architecture/engine.md, and none. A slot of any content takes a role,
// so a plan can give roles before chords.
func TestWithRole(t *testing.T) {
	roles := []progression.Role{
		progression.RoleNone,
		progression.RoleStable, progression.RoleLink, progression.RolePrepare,
		progression.RoleRise, progression.RoleResolve, progression.RoleClose,
		progression.RoleSpread, progression.RoleFloat, progression.RoleWistful, progression.RoleLinger,
	}
	seen := map[progression.Role]bool{}
	for _, r := range roles {
		if seen[r] {
			t.Errorf("%v appears twice; want eleven different values", r)
		}
		seen[r] = true
		for _, c := range []progression.Content{progression.Unchosen(), progression.NoChord(), progression.ChordOf(material.Chord{})} {
			s, err := slot(c, 4).WithRole(r)
			if err != nil || s.Role() != r || s.Beats() != 4 || s.Content() != c {
				t.Errorf("WithRole(%v) = %v, %v; want the role, the length and the content kept", r, s, err)
			}
		}
	}
	for _, r := range []progression.Role{progression.RoleLinger + 1, progression.RoleNone - 1} {
		if _, err := slot(progression.NoChord(), 4).WithRole(r); !errors.Is(err, progression.ErrRole) {
			t.Errorf("WithRole(%v): err = %v; want ErrRole", r, err)
		}
	}
}

func TestWithContent(t *testing.T) {
	c := progression.ChordOf(material.Chord{})
	contents := []progression.Content{progression.Unchosen(), progression.NoChord(), c}
	for _, from := range contents {
		for _, to := range contents {
			s, err := slot(from, 6).WithRole(progression.RoleRise)
			if err != nil {
				t.Fatal(err)
			}
			got := s.WithContent(to)
			if got.Content() != to || got.Beats() != 6 || got.Role() != progression.RoleRise {
				t.Errorf("WithContent(%v) on %v = %v; want the content replaced, length and role kept", to, from, got)
			}
			if s.Content() != from {
				t.Errorf("WithContent changed the original slot to %v", s.Content())
			}
		}
	}
}
