package material

import "fmt"

// Transpose returns the note name the interval above n. The letter moves by
// the interval number, so the spelling is kept: E up a major third is G♯, not
// A♭. The diminished unison is the one interval that moves down (C to C♭).
// It fails when the result needs more than a double sharp or a double flat.
func (n NoteName) Transpose(iv Interval) (NoteName, error) {
	return n.moveBy(iv.steps, iv.Semitones(), iv, "up")
}

// TransposeDown returns the note name the interval below n, keeping the
// spelling in the same way as Transpose.
func (n NoteName) TransposeDown(iv Interval) (NoteName, error) {
	return n.moveBy(-iv.steps, -iv.Semitones(), iv, "down")
}

func (n NoteName) moveBy(steps, semitones int, iv Interval, direction string) (NoteName, error) {
	position := int(n.letter) + steps
	alter := n.semitones() + semitones - naturalAt(position)
	if alter < minAlter || alter > maxAlter {
		return NoteName{}, fmt.Errorf("%s %s %s: %w", n, direction, iv, ErrAlterOutOfRange)
	}
	return NoteName{letter: Letter(mod(position, letters)), alter: alter}, nil
}
