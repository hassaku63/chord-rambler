package progression

import (
	"errors"
	"fmt"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
)

// ErrLength reports a slot length of less than one beat.
var ErrLength = errors.New("slot length is less than one beat")

// ErrRole reports a role that is not one of the ten roles or none.
var ErrRole = errors.New("no such role")

type kind int

const (
	unchosen kind = iota
	noChord
	hasChord
)

// Content is what a slot holds: a chord, no chord (nothing sounds), or none
// chosen yet. The three are different values. The zero value is none chosen
// yet. Build the others with NoChord and ChordOf.
type Content struct {
	kind  kind
	chord material.Chord // set only for hasChord
}

// Unchosen returns the content of a slot whose chord is not chosen yet: the
// zero value.
func Unchosen() Content { return Content{} }

// NoChord returns the content of a slot that sounds nothing.
func NoChord() Content { return Content{kind: noChord} }

// ChordOf returns the content of a slot that sounds the chord.
func ChordOf(c material.Chord) Content { return Content{kind: hasChord, chord: c} }

// Chord returns the chord and true, or false when the content is no chord or
// none chosen yet.
func (c Content) Chord() (material.Chord, bool) { return c.chord, c.kind == hasChord }

// IsNoChord reports whether the content is no chord: nothing sounds.
func (c Content) IsNoChord() bool { return c.kind == noChord }

// IsUnchosen reports whether no chord has been chosen yet.
func (c Content) IsUnchosen() bool { return c.kind == unchosen }

// Role is what a slot does in the progression: one of the six structural
// roles and the four colour roles of docs/architecture/engine.md, or none.
// Slot.WithRole checks the range.
type Role int

// The roles. RoleNone, the zero value, is no role.
const (
	RoleNone Role = iota
	// The structural roles.
	RoleStable  // 安定
	RoleLink    // つなぎ
	RolePrepare // 準備
	RoleRise    // 高まり
	RoleResolve // 解決
	RoleClose   // 締め
	// The colour roles.
	RoleSpread  // 広がり
	RoleFloat   // 浮遊
	RoleWistful // 切なさ
	RoleLinger  // 余韻
)

var roleNames = [...]string{
	"RoleNone",
	"RoleStable", "RoleLink", "RolePrepare", "RoleRise", "RoleResolve", "RoleClose",
	"RoleSpread", "RoleFloat", "RoleWistful", "RoleLinger",
}

func (r Role) valid() bool { return r >= RoleNone && int(r) < len(roleNames) }

// String returns the name of the constant, such as "RoleStable", for tests and
// logs.
func (r Role) String() string {
	if !r.valid() {
		return fmt.Sprintf("Role(%d)", int(r))
	}
	return roleNames[r]
}

// Slot is one place in a progression: its content, its length in whole beats
// (one or more) and an optional role. The zero value is none chosen yet, one
// beat long, with no role. Build other values with NewSlot and NewBarSlot.
type Slot struct {
	content    Content
	extraBeats int // the length minus one, so that the zero value is one beat
	role       Role
}

// NewSlot returns a slot of the content, the beats long, with no role. It fails
// for a length of less than one beat (ErrLength).
func NewSlot(c Content, beats int) (Slot, error) {
	if beats < 1 {
		return Slot{}, fmt.Errorf("slot of %d beats: %w", beats, ErrLength)
	}
	return Slot{content: c, extraBeats: beats - 1}, nil
}

// NewBarSlot returns a slot of the content one bar of the meter long, the
// default length, with no role.
func NewBarSlot(c Content, m Meter) Slot {
	return Slot{content: c, extraBeats: m.BeatsPerBar() - 1}
}

// Content returns what the slot holds.
func (s Slot) Content() Content { return s.content }

// Beats returns the length in beats.
func (s Slot) Beats() int { return s.extraBeats + 1 }

// Role returns the role, or RoleNone.
func (s Slot) Role() Role { return s.role }

// WithRole returns the slot with the role. RoleNone removes the role. It fails
// for a value that is not one of the constants (ErrRole).
func (s Slot) WithRole(r Role) (Slot, error) {
	if !r.valid() {
		return Slot{}, fmt.Errorf("%s: %w", r, ErrRole)
	}
	s.role = r
	return s, nil
}
