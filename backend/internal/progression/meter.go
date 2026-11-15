package progression

import (
	"errors"
	"fmt"
)

// ErrMeter reports a meter that cannot be built yet. The first scope takes 4/4
// only (#91).
var ErrMeter = errors.New("meter is not 4/4")

// Meter is the time signature of a progression. Only 4/4 can be built now, so
// the zero value is 4/4 and so is every value. Build it with NewMeter, or use
// the zero value.
type Meter struct{}

// NewMeter returns the meter of the beats in a bar over the note value of a
// beat (4 and 4 for 4/4). It fails for anything other than 4/4 (ErrMeter).
func NewMeter(beats, unit int) (Meter, error) {
	if beats != 4 || unit != 4 {
		return Meter{}, fmt.Errorf("meter %d/%d: %w", beats, unit, ErrMeter)
	}
	return Meter{}, nil
}

// BeatsPerBar returns the number of beats in a bar.
func (Meter) BeatsPerBar() int { return 4 }
