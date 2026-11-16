package reading

import (
	"fmt"

	"github.com/hassaku63/chord-rambler/backend/internal/progression"
	"github.com/hassaku63/chord-rambler/backend/internal/relation"
)

// RoleFit is whether the reading of a slot fits the structural role put on it
// (docs/architecture/theory-conventions.md, 7). The zero value, RoleNotJudged,
// says nothing.
type RoleFit int

// The verdicts.
const (
	// RoleNotJudged is said of a slot with no role or a colour role.
	RoleNotJudged RoleFit = iota
	// RoleFits is a slot whose reading meets the rule of its role.
	RoleFits
	// RoleMisfits is a slot whose reading does not.
	RoleMisfits
	// RoleUndecided is a slot whose verdict depends on a chord not there: its
	// own, or that of the slot the role looks at.
	RoleUndecided
)

// String returns the name of the constant, for tests and logs.
func (f RoleFit) String() string {
	switch f {
	case RoleNotJudged:
		return "RoleNotJudged"
	case RoleFits:
		return "RoleFits"
	case RoleMisfits:
		return "RoleMisfits"
	case RoleUndecided:
		return "RoleUndecided"
	default:
		return fmt.Sprintf("RoleFit(%d)", int(f))
	}
}

// ReadRoleFits returns, for every slot of the progression in order (as
// ReadProgression numbers them), whether its reading fits its structural role.
// It looks only at the readings: the main function of the chord (ReadChord),
// whether the chord is tertian (Tertian), the resolution and the cadence
// (ReadProgression) and whether the phrase is full (relation.Places). A
// function counts only on a tertian chord, so Gsus4 and G6 are not dominants
// here, and a chord with no function (borrowed, a secondary dominant, facts
// only) is of no function. The previous and next slots are those of the
// progression, across phrases.
//
// The verdict is decided in this order:
//
//  1. no verdict for a slot with no role or a colour role (the owner's call
//     C1 on #124);
//  2. undecided for a slot with no chord or none chosen yet;
//  3. undecided when the slot the role looks at has no chord, none chosen
//     yet, or is not there: the next for 準備, the previous for つなぎ and
//     解決 (B1 on #124). For 締め, undecided on the last slot of a phrase that
//     is not full;
//  4. fits when the rule of the role is met, and misfits otherwise:
//     - 安定: tonic.
//     - 高まり: dominant.
//     - 準備: subdominant, and the next slot dominant.
//     - つなぎ: the function of the previous slot, on another root (C after
//     C/E, Cmaj7 after C and G7 after G do not fit).
//     - 解決: the previous slot is V and resolves as expected (to I).
//     - 締め: the slot is the last of its phrase and a cadence is named on it.
//
// It fails for a key other than Ionian (relation.ErrNotMajor).
func ReadRoleFits(p progression.Progression) ([]RoleFit, error) {
	read, err := ReadProgression(p)
	if err != nil {
		return nil, fmt.Errorf("read role fits: %w", err)
	}
	v := roleView{slots: p.Slots(), read: read, places: relation.Places(p)}
	out := make([]RoleFit, len(v.slots))
	for i, s := range v.slots {
		out[i] = v.fit(i, s.Role())
	}
	return out, nil
}

// roleView is what the rules of the roles look at: the slots, their readings
// and their places, all in the order of the progression.
type roleView struct {
	slots  []progression.Slot
	read   []SlotReading
	places []relation.Place
}

// missing reports whether the slot at i holds no chord or is not there.
func (v roleView) missing(i int) bool {
	return i < 0 || i >= len(v.slots) || !v.read[i].HasChord
}

// function returns the function of the chord of the slot at i when it is
// tertian, and NoFunction otherwise or when the slot holds no chord.
func (v roleView) function(i int) Function {
	if v.missing(i) {
		return NoFunction
	}
	if c, _ := v.slots[i].Content().Chord(); !Tertian(c) {
		return NoFunction
	}
	return v.read[i].Chord.Main.Function
}

// fit returns the verdict on the slot at i for the role.
func (v roleView) fit(i int, role progression.Role) RoleFit {
	rule, ok := structuralRules[role]
	switch {
	case !ok:
		return RoleNotJudged // no role, or a colour role
	case v.missing(i):
		return RoleUndecided
	default:
		return rule(v, i)
	}
}

// structuralRules holds the rule of each structural role, for a slot that
// holds a chord.
var structuralRules = map[progression.Role]func(v roleView, i int) RoleFit{
	progression.RoleStable: func(v roleView, i int) RoleFit {
		return verdict(v.function(i) == Tonic)
	},
	progression.RoleRise: func(v roleView, i int) RoleFit {
		return verdict(v.function(i) == Dominant)
	},
	progression.RolePrepare: func(v roleView, i int) RoleFit {
		if v.missing(i + 1) {
			return RoleUndecided
		}
		return verdict(v.function(i) == Subdominant && v.function(i+1) == Dominant)
	},
	progression.RoleLink: func(v roleView, i int) RoleFit {
		if v.missing(i - 1) {
			return RoleUndecided
		}
		cur, _ := v.slots[i].Content().Chord()
		prev, _ := v.slots[i-1].Content().Chord()
		return verdict(v.function(i) != NoFunction && v.function(i) == v.function(i-1) && cur.Root() != prev.Root())
	},
	progression.RoleResolve: func(v roleView, i int) RoleFit {
		if v.missing(i - 1) {
			return RoleUndecided
		}
		prev := v.read[i-1]
		return verdict(prev.Resolution == ResolutionExpected && prev.Chord.Main.Kind == Diatonic)
	},
	progression.RoleClose: func(v roleView, i int) RoleFit {
		place := v.places[i]
		if place.LastInPhrase && !place.FullPhrase {
			return RoleUndecided
		}
		return verdict(place.LastInPhrase && v.read[i].Cadence != NoCadence)
	},
}

func verdict(fits bool) RoleFit {
	if fits {
		return RoleFits
	}
	return RoleMisfits
}
