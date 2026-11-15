package material

import (
	"errors"
	"fmt"
	"strconv"
)

// ErrInvalidInterval reports an interval that cannot be named: a number
// outside 1..13, an unknown quality, a quality that the number does not take
// (a major fifth, a perfect third), or two note names whose distance has none
// of the five qualities.
var ErrInvalidInterval = errors.New("no such interval")

// The interval numbers run from a unison (1) to a thirteenth (13), so that the
// tensions 9, 11 and 13 have their own numbers.
const (
	minNumber = 1
	maxNumber = 13
)

// Quality is the kind of an interval. Perfect is the zero value.
type Quality int

// The five qualities.
const (
	Perfect Quality = iota
	Major
	Minor
	Augmented
	Diminished
)

// String returns the abbreviation of the quality: P, M, m, A or d.
func (q Quality) String() string {
	switch q {
	case Perfect:
		return "P"
	case Major:
		return "M"
	case Minor:
		return "m"
	case Augmented:
		return "A"
	case Diminished:
		return "d"
	default:
		return fmt.Sprintf("Quality(%d)", int(q))
	}
}

// Interval is a spelled interval: a number (1 to 13) and a quality. The zero
// value is a perfect unison. Build other values with NewInterval or
// IntervalBetween.
type Interval struct {
	steps   int // the number minus one: letters moved
	quality Quality
}

// NewInterval returns the interval with the number and the quality. Unisons,
// fourths, fifths, octaves, elevenths and twelfths take perfect, augmented and
// diminished; the other numbers take major, minor, augmented and diminished.
func NewInterval(number int, quality Quality) (Interval, error) {
	if number < minNumber || number > maxNumber {
		return Interval{}, fmt.Errorf("interval number %d: %w", number, ErrInvalidInterval)
	}
	iv := Interval{steps: number - 1, quality: quality}
	if _, ok := qualityOffset(iv.perfectKind(), quality); !ok {
		return Interval{}, fmt.Errorf("interval %s: %w", iv, ErrInvalidInterval)
	}
	return iv, nil
}

// IntervalBetween returns the simple interval (1 to 7) from the lower note
// name up to the upper one. The number comes from the letters, so C to D♯ is
// an augmented second and C to E♭ a minor third. On the same letter the
// "upper" note may sound lower: C♯ to C is a diminished unison. It fails when
// the distance has none of the five qualities (C♯ to E♭♭).
func IntervalBetween(lower, upper NoteName) (Interval, error) {
	steps := mod(int(upper.letter)-int(lower.letter), letters)
	semitones := naturalAt(int(lower.letter)+steps) + upper.alter - lower.semitones()
	iv := Interval{steps: steps}
	offset := semitones - naturalAt(steps)
	for _, q := range []Quality{Perfect, Major, Minor, Augmented, Diminished} {
		if o, ok := qualityOffset(iv.perfectKind(), q); ok && o == offset {
			iv.quality = q
			return iv, nil
		}
	}
	return Interval{}, fmt.Errorf("interval from %s up to %s: %w", lower, upper, ErrInvalidInterval)
}

// Number returns the interval number, 1 to 13.
func (iv Interval) Number() int { return iv.steps + 1 }

// Quality returns the quality.
func (iv Interval) Quality() Quality { return iv.quality }

// Semitones returns the size of the interval in semitones. A diminished unison
// is -1.
func (iv Interval) Semitones() int {
	offset, _ := qualityOffset(iv.perfectKind(), iv.quality)
	return naturalAt(iv.steps) + offset
}

// String returns the interval for reading, such as "M3" or "A4".
func (iv Interval) String() string {
	return iv.quality.String() + strconv.Itoa(iv.Number())
}

// perfectKind reports whether the number takes perfect rather than major and
// minor: unisons, fourths and fifths, and the same an octave higher.
func (iv Interval) perfectKind() bool {
	switch iv.steps % letters {
	case 0, 3, 4:
		return true
	default:
		return false
	}
}

// qualityOffset returns how many semitones the quality adds to the perfect or
// major interval of the same number, and whether the number takes the quality.
func qualityOffset(perfectKind bool, q Quality) (int, bool) {
	switch q {
	case Perfect:
		return 0, perfectKind
	case Major:
		return 0, !perfectKind
	case Minor:
		return -1, !perfectKind
	case Augmented:
		return 1, true
	case Diminished:
		if perfectKind {
			return -1, true
		}
		return -2, true
	default:
		return 0, false
	}
}
