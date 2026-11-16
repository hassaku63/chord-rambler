package reading_test

import (
	"errors"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/progression"
	"github.com/hassaku63/chord-rambler/backend/internal/reading"
	"github.com/hassaku63/chord-rambler/backend/internal/relation"
)

// roled is a slot of one bar of 4/4 of a test progression, with a role.
type roled struct {
	content progression.Content
	role    progression.Role
}

// on is a bar holding the chord, with the role; plain is one with no role.
func on(c material.Chord, r progression.Role) roled { return roled{progression.ChordOf(c), r} }
func plain(c material.Chord) roled                  { return roled{content: progression.ChordOf(c)} }

// buildRoled appends the slots to an empty progression in the key, in 4/4.
func buildRoled(key material.Scale, slots ...roled) progression.Progression {
	p := progression.New(key, progression.Meter{})
	for _, s := range slots {
		ps, err := progression.NewSlot(s.content, 4)
		if err == nil {
			ps, err = ps.WithRole(s.role)
		}
		if err == nil {
			p, err = p.Append(ps)
		}
		if err != nil {
			panic(err)
		}
	}
	return p
}

const (
	stable  = progression.RoleStable
	link    = progression.RoleLink
	prepare = progression.RolePrepare
	rise    = progression.RoleRise
	resolve = progression.RoleResolve
	closing = progression.RoleClose
)

var (
	aFlatMaj  = chord(aFlat, maj, "")
	f6        = chord(fNat, "M3 P5 M6", "")
	noneRoled = func(r progression.Role) roled { return roled{progression.NoChord(), r} }
	unRoled   = func(r progression.Role) roled { return roled{progression.Unchosen(), r} }
)

// roleCases are progressions in C major, each with the slot judged (at) and
// the verdict expected. The rules are those of #125, said again from the
// readings (docs/architecture/theory-conventions.md, 7): the function counts
// only on a tertian chord; つなぎ needs another root; 解決 needs V resolving as
// expected; 締め needs a cadence on the last slot of a full phrase. Four bars
// make a full phrase.
var roleCases = []struct {
	name  string
	slots []roled
	at    int
	want  reading.RoleFit
}{
	// 安定
	{"安定: C", []roled{on(c, stable)}, 0, reading.RoleFits},
	{"安定: Am (vi is tonic)", []roled{on(am, stable)}, 0, reading.RoleFits},
	{"安定: C/E", []roled{on(cOverE, stable)}, 0, reading.RoleFits},
	{"安定: G", []roled{on(g, stable)}, 0, reading.RoleMisfits},
	{"安定: C6 (no shape)", []roled{on(c6, stable)}, 0, reading.RoleMisfits},
	{"安定: Fm (borrowed, no function)", []roled{on(fm, stable)}, 0, reading.RoleMisfits},
	{"安定: none chosen", []roled{unRoled(stable)}, 0, reading.RoleUndecided},
	{"安定: no chord", []roled{noneRoled(stable)}, 0, reading.RoleUndecided},
	// 高まり
	{"高まり: G7", []roled{on(g7, rise)}, 0, reading.RoleFits},
	{"高まり: Bdim (vii° is dominant)", []roled{on(bdim, rise)}, 0, reading.RoleFits},
	{"高まり: C", []roled{on(c, rise)}, 0, reading.RoleMisfits},
	{"高まり: Gsus4", []roled{on(gsus4, rise)}, 0, reading.RoleMisfits},
	{"高まり: G6", []roled{on(g6, rise)}, 0, reading.RoleMisfits},
	{"高まり: D7 (secondary dominant, no function)", []roled{on(d7, rise)}, 0, reading.RoleMisfits},
	{"高まり: none chosen", []roled{unRoled(rise)}, 0, reading.RoleUndecided},
	{"高まり: no chord", []roled{noneRoled(rise)}, 0, reading.RoleUndecided},
	// 準備
	{"準備: F before G", []roled{on(f, prepare), plain(g)}, 0, reading.RoleFits},
	{"準備: Dm before G7", []roled{on(dm, prepare), plain(g7)}, 0, reading.RoleFits},
	{"準備: F before C", []roled{on(f, prepare), plain(c)}, 0, reading.RoleMisfits},
	{"準備: C before G", []roled{on(c, prepare), plain(g)}, 0, reading.RoleMisfits},
	{"準備: F6 before G", []roled{on(f6, prepare), plain(g)}, 0, reading.RoleMisfits},
	{"準備: F before Gsus4", []roled{on(f, prepare), plain(gsus4)}, 0, reading.RoleMisfits},
	{"準備: F before D7", []roled{on(f, prepare), plain(d7)}, 0, reading.RoleMisfits},
	{"準備: F last", []roled{plain(c), on(f, prepare)}, 1, reading.RoleUndecided},
	{"準備: F before none chosen", []roled{on(f, prepare), unRoled(0)}, 0, reading.RoleUndecided},
	{"準備: F before no chord", []roled{on(f, prepare), noneRoled(0)}, 0, reading.RoleUndecided},
	{"準備: C before none chosen", []roled{on(c, prepare), unRoled(0)}, 0, reading.RoleUndecided},
	{"準備: none chosen", []roled{unRoled(prepare), plain(g)}, 0, reading.RoleUndecided},
	// つなぎ
	{"つなぎ: Am after C", []roled{plain(c), on(am, link)}, 1, reading.RoleFits},
	{"つなぎ: Dm after F", []roled{plain(f), on(dm, link)}, 1, reading.RoleFits},
	{"つなぎ: Am after C, across phrases",
		[]roled{plain(c), plain(f), plain(g7), plain(c), on(am, link)}, 4, reading.RoleFits},
	{"つなぎ: C after C", []roled{plain(c), on(c, link)}, 1, reading.RoleMisfits},
	{"つなぎ: C/E after C", []roled{plain(c), on(cOverE, link)}, 1, reading.RoleMisfits},
	{"つなぎ: Cmaj7 after C", []roled{plain(c), on(cMaj7, link)}, 1, reading.RoleMisfits},
	{"つなぎ: G7 after G", []roled{plain(g), on(g7, link)}, 1, reading.RoleMisfits},
	{"つなぎ: F after C", []roled{plain(c), on(f, link)}, 1, reading.RoleMisfits},
	{"つなぎ: A♭ after Fm (no function)", []roled{plain(fm), on(aFlatMaj, link)}, 1, reading.RoleMisfits},
	{"つなぎ: Bdim after Gsus4", []roled{plain(gsus4), on(bdim, link)}, 1, reading.RoleMisfits},
	{"つなぎ: C6 after Am", []roled{plain(am), on(c6, link)}, 1, reading.RoleMisfits},
	{"つなぎ: first", []roled{on(am, link)}, 0, reading.RoleUndecided},
	{"つなぎ: after none chosen", []roled{unRoled(0), on(am, link)}, 1, reading.RoleUndecided},
	{"つなぎ: after no chord", []roled{noneRoled(0), on(am, link)}, 1, reading.RoleUndecided},
	// 解決
	{"解決: C after G7", []roled{plain(g7), on(c, resolve)}, 1, reading.RoleFits},
	{"解決: C/E after G", []roled{plain(g), on(cOverE, resolve)}, 1, reading.RoleFits},
	{"解決: C after G and G7", []roled{plain(g), plain(g7), on(c, resolve)}, 2, reading.RoleFits},
	{"解決: Am after G7", []roled{plain(g7), on(am, resolve)}, 1, reading.RoleMisfits},
	{"解決: C after Bdim (not V)", []roled{plain(bdim), on(c, resolve)}, 1, reading.RoleMisfits},
	{"解決: G after D7 (not to I)", []roled{plain(d7), on(g, resolve)}, 1, reading.RoleMisfits},
	{"解決: C after F", []roled{plain(f), on(c, resolve)}, 1, reading.RoleMisfits},
	{"解決: G7 after G", []roled{plain(g), on(g7, resolve)}, 1, reading.RoleMisfits},
	{"解決: Cm after G7", []roled{plain(g7), on(cm, resolve)}, 1, reading.RoleMisfits},
	{"解決: C6 after G7", []roled{plain(g7), on(c6, resolve)}, 1, reading.RoleMisfits},
	{"解決: first", []roled{on(c, resolve)}, 0, reading.RoleUndecided},
	{"解決: after none chosen", []roled{unRoled(0), on(c, resolve)}, 1, reading.RoleUndecided},
	{"解決: after no chord", []roled{noneRoled(0), on(c, resolve)}, 1, reading.RoleUndecided},
	// 締め
	{"締め: authentic", []roled{plain(c), plain(f), plain(g7), on(c, closing)}, 3, reading.RoleFits},
	{"締め: half", []roled{plain(c), plain(f), plain(dm), on(g, closing)}, 3, reading.RoleFits},
	{"締め: deceptive", []roled{plain(c), plain(f), plain(g7), on(am, closing)}, 3, reading.RoleFits},
	{"締め: plagal", []roled{plain(c), plain(am), plain(f), on(c, closing)}, 3, reading.RoleFits},
	{"締め: no cadence", []roled{plain(c), plain(g), plain(am), on(f, closing)}, 3, reading.RoleMisfits},
	{"締め: not last", []roled{plain(c), plain(f), on(g7, closing), plain(c)}, 2, reading.RoleMisfits},
	{"締め: not last of a phrase not full", []roled{on(c, closing), plain(f)}, 0, reading.RoleMisfits},
	{"締め: after none chosen",
		[]roled{plain(c), plain(f), unRoled(0), on(c, closing)}, 3, reading.RoleMisfits},
	{"締め: phrase not full", []roled{plain(f), plain(g7), on(c, closing)}, 2, reading.RoleUndecided},
	{"締め: none chosen, last", []roled{plain(c), plain(f), plain(g7), unRoled(closing)}, 3, reading.RoleUndecided},
	{"締め: none chosen, not last", []roled{unRoled(closing), plain(c)}, 0, reading.RoleUndecided},
	// No role, and the colour roles (C1 on #124).
	{"no role", []roled{plain(c)}, 0, reading.RoleNotJudged},
	{"広がり", []roled{on(c, progression.RoleSpread)}, 0, reading.RoleNotJudged},
	{"余韻 on none chosen", []roled{unRoled(progression.RoleLinger)}, 0, reading.RoleNotJudged},
}

func TestRoleFits(t *testing.T) {
	for _, tt := range roleCases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := reading.ReadRoleFits(buildRoled(major(cNat), tt.slots...))
			if err != nil {
				t.Fatalf("ReadRoleFits: %v", err)
			}
			if len(got) != len(tt.slots) {
				t.Fatalf("%d verdicts for %d slots", len(got), len(tt.slots))
			}
			if got[tt.at] != tt.want {
				t.Errorf("slot %d: %s; want %s", tt.at, got[tt.at], tt.want)
			}
		})
	}
}

// Each progression of roleCases, moved to the twelve major keys, gets the same
// verdict: the rules look at functions, shapes and readings only.
func TestRoleFitsInMajorKeys(t *testing.T) {
	intervals := []string{"P1", "P5", "M2", "M6", "M3", "M7", "A4", "m2", "m6", "m3", "m7", "P4"}
	for _, s := range intervals {
		iv := ivs(s)[0]
		for _, tt := range roleCases {
			p, err := buildRoled(major(cNat), tt.slots...).Transpose(iv)
			if err != nil {
				t.Fatalf("%s up %s: %v", tt.name, s, err)
			}
			got, err := reading.ReadRoleFits(p)
			if err != nil {
				t.Fatalf("%s in %s: %v", tt.name, p.Key().Tonic(), err)
			}
			if got[tt.at] != tt.want {
				t.Errorf("%s in %s: %s; want %s", tt.name, p.Key().Tonic(), got[tt.at], tt.want)
			}
		}
	}
}

func TestRoleFitsNotMajor(t *testing.T) {
	minor, err := material.NewScale(aNat, material.Aeolian)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := reading.ReadRoleFits(buildRoled(minor, on(am, stable))); !errors.Is(err, relation.ErrNotMajor) {
		t.Errorf("ReadRoleFits in A minor: %v; want relation.ErrNotMajor", err)
	}
}
