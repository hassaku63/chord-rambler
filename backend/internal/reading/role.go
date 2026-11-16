package reading

import (
	"fmt"
	"slices"

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

// RoleCondition is one condition of the rule of a structural role
// (docs/architecture/theory-conventions.md, 7). A slot that does not fit its
// role is told which of them it does not meet.
type RoleCondition int

// The conditions, in the order they are listed.
const (
	// ConditionShape is that the chord of the slot is a triad with a third or
	// a seventh chord (Tertian).
	ConditionShape RoleCondition = iota
	// ConditionFunction is that the main function of the chord of the slot is
	// that of the role: for つなぎ, a function, and that of the previous slot.
	ConditionFunction
	// ConditionPreviousV is that the previous slot is V, as the resolution
	// and the cadence read V.
	ConditionPreviousV
	// ConditionNextDominant is that the next slot is a dominant.
	ConditionNextDominant
	// ConditionCadence is that a cadence is named on the slot, so it is also
	// that the slot is the last of its phrase.
	ConditionCadence
	// ConditionAnotherRoot is that the root of the chord is not that of the
	// previous slot.
	ConditionAnotherRoot
	// ConditionExpectedResolution is that V on the previous slot resolves as
	// expected, to I on the slot. It is not looked at when the previous slot
	// is not V.
	ConditionExpectedResolution
)

// String returns the name of the constant, for tests and logs.
func (c RoleCondition) String() string {
	switch c {
	case ConditionShape:
		return "ConditionShape"
	case ConditionFunction:
		return "ConditionFunction"
	case ConditionPreviousV:
		return "ConditionPreviousV"
	case ConditionNextDominant:
		return "ConditionNextDominant"
	case ConditionCadence:
		return "ConditionCadence"
	case ConditionAnotherRoot:
		return "ConditionAnotherRoot"
	case ConditionExpectedResolution:
		return "ConditionExpectedResolution"
	default:
		return fmt.Sprintf("RoleCondition(%d)", int(c))
	}
}

// RoleVerdict is whether the reading of a slot fits its structural role, and
// why not when it does not.
type RoleVerdict struct {
	Fit RoleFit
	// Unmet holds the conditions of the rule the slot does not meet, in the
	// order of the constants, when Fit is RoleMisfits, and is nil otherwise.
	Unmet []RoleCondition
}

// ReadRoleFits returns, for every slot of the progression in order, the Fit of
// ReadRoleVerdicts.
//
// It fails for a key other than Ionian (relation.ErrNotMajor).
func ReadRoleFits(p progression.Progression) ([]RoleFit, error) {
	verdicts, err := ReadRoleVerdicts(p)
	if err != nil {
		return nil, err
	}
	out := make([]RoleFit, len(verdicts))
	for i, v := range verdicts {
		out[i] = v.Fit
	}
	return out, nil
}

// ReadRoleVerdicts returns, for every slot of the progression in order (as
// ReadProgression numbers them), whether its reading fits its structural role
// and, when it does not, the conditions it does not meet. It looks only at the
// readings: the main function of the chord (ReadChord), whether the chord is
// tertian (Tertian), whether a chord is V, the resolution and the cadence
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
//     is not full, and, when no cadence is named on the last slot of a full
//     phrase, when the slot before it in the phrase has no chord, none chosen
//     yet, or is not there (a half cadence needs no slot before it);
//  4. fits when every condition of the rule of the role is met, and misfits
//     with those not met otherwise:
//     - 安定: shape, and the function tonic.
//     - 高まり: shape, and the function dominant.
//     - 準備: shape, the function subdominant, and the next slot dominant.
//     - つなぎ: shape, the function of the previous slot, and another root
//     (C after C/E, Cmaj7 after C and G7 after G do not fit).
//     - 解決: the previous slot is V, and resolves as expected (to I).
//     - 締め: a cadence is named on the slot, so it is the last of its phrase.
//
// It fails for a key other than Ionian (relation.ErrNotMajor).
func ReadRoleVerdicts(p progression.Progression) ([]RoleVerdict, error) {
	read, err := ReadProgression(p)
	if err != nil {
		return nil, fmt.Errorf("read role fits: %w", err)
	}
	chords, err := readSlots(p)
	if err != nil {
		return nil, fmt.Errorf("read role fits: %w", err)
	}
	v := roleView{slots: p.Slots(), read: read, chords: chords, places: relation.Places(p)}
	out := make([]RoleVerdict, len(v.slots))
	for i, s := range v.slots {
		out[i] = v.verdict(i, s.Role())
	}
	return out, nil
}

// roleView is what the rules of the roles look at: the slots, their readings,
// what the rules of ReadProgression look at in them, and their places, all in
// the order of the progression.
type roleView struct {
	slots  []progression.Slot
	read   []SlotReading
	chords []slotChord
	places []relation.Place
}

// missing reports whether the slot at i holds no chord or is not there.
func (v roleView) missing(i int) bool {
	return i < 0 || i >= len(v.slots) || !v.read[i].HasChord
}

// function returns the function of the chord of the slot at i when it is
// tertian, and NoFunction otherwise or when the slot holds no chord.
func (v roleView) function(i int) Function {
	if v.missing(i) || !v.chords[i].tertian {
		return NoFunction
	}
	return v.read[i].Chord.Main.Function
}

// verdict returns the verdict on the slot at i for the role.
func (v roleView) verdict(i int, role progression.Role) RoleVerdict {
	rule, ok := structuralRules[role]
	switch {
	case !ok:
		return RoleVerdict{} // no role, or a colour role
	case v.missing(i):
		return RoleVerdict{Fit: RoleUndecided}
	}
	undecided, unmet := rule(v, i)
	switch {
	case undecided:
		return RoleVerdict{Fit: RoleUndecided}
	case len(unmet) == 0:
		return RoleVerdict{Fit: RoleFits}
	default:
		slices.Sort(unmet)
		return RoleVerdict{Fit: RoleMisfits, Unmet: unmet}
	}
}

// unmet collects the conditions not met.
type unmet []RoleCondition

// unless adds the condition when it is not met.
func (u unmet) unless(met bool, c RoleCondition) unmet {
	if met {
		return u
	}
	return append(u, c)
}

// shaped returns the conditions of shape and function not met by the chord of
// the slot at i, for the function want.
func (v roleView) shaped(i int, want Function) unmet {
	return unmet(nil).
		unless(v.chords[i].tertian, ConditionShape).
		unless(v.read[i].Chord.Main.Function == want, ConditionFunction)
}

// structuralRules holds the rule of each structural role, for a slot that
// holds a chord: whether the verdict is undecided, and otherwise the
// conditions not met.
var structuralRules = map[progression.Role]func(v roleView, i int) (bool, unmet){
	progression.RoleStable: func(v roleView, i int) (bool, unmet) {
		return false, v.shaped(i, Tonic)
	},
	progression.RoleRise: func(v roleView, i int) (bool, unmet) {
		return false, v.shaped(i, Dominant)
	},
	progression.RolePrepare: func(v roleView, i int) (bool, unmet) {
		if v.missing(i + 1) {
			return true, nil
		}
		return false, v.shaped(i, Subdominant).unless(v.function(i+1) == Dominant, ConditionNextDominant)
	},
	progression.RoleLink: func(v roleView, i int) (bool, unmet) {
		if v.missing(i - 1) {
			return true, nil
		}
		cur, _ := v.slots[i].Content().Chord()
		prev, _ := v.slots[i-1].Content().Chord()
		f := v.read[i].Chord.Main.Function
		return false, unmet(nil).
			unless(v.chords[i].tertian, ConditionShape).
			unless(f != NoFunction && f == v.function(i-1), ConditionFunction).
			unless(cur.Root() != prev.Root(), ConditionAnotherRoot)
	},
	progression.RoleResolve: func(v roleView, i int) (bool, unmet) {
		if v.missing(i - 1) {
			return true, nil
		}
		isV := v.chords[i-1].v
		return false, unmet(nil).
			unless(isV, ConditionPreviousV).
			unless(!isV || v.read[i-1].Resolution == ResolutionExpected, ConditionExpectedResolution)
	},
	progression.RoleClose: func(v roleView, i int) (bool, unmet) {
		place := v.places[i]
		switch {
		case place.LastInPhrase && !place.FullPhrase:
			return true, nil
		case !place.LastInPhrase:
			return false, unmet{ConditionCadence}
		case v.read[i].Cadence != NoCadence:
			return false, nil
		case place.IndexInPhrase == 1 || v.missing(i-1):
			return true, nil
		default:
			return false, unmet{ConditionCadence}
		}
	},
}
