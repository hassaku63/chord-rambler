package material

import "fmt"

// Triad is the view of a chord from its third and fifth, or, without a third,
// from the fourth or second that replaces it. TriadOther is the zero value.
type Triad int

// The triad views.
const (
	// TriadOther is any chord outside the views below, such as one with a
	// fifth and no third, second or fourth.
	TriadOther      Triad = iota
	TriadMajor            // major third, perfect fifth
	TriadMinor            // minor third, perfect fifth
	TriadDiminished       // minor third, diminished fifth
	TriadAugmented        // major third, augmented fifth
	TriadSus4             // no third; perfect fourth, perfect fifth
	TriadSus2             // no third; major second, perfect fifth
)

// String returns the name of the constant, such as "TriadSus4", for tests and
// logs. Chord symbols belong to the notation package.
func (t Triad) String() string {
	names := [...]string{"TriadOther", "TriadMajor", "TriadMinor", "TriadDiminished", "TriadAugmented", "TriadSus4", "TriadSus2"}
	if t < 0 || int(t) >= len(names) {
		return fmt.Sprintf("Triad(%d)", int(t))
	}
	return names[t]
}

// Seventh is the view of a chord from its seventh. NoSeventh is the zero value.
type Seventh int

// The seventh views.
const (
	NoSeventh         Seventh = iota
	SeventhMajor              // major seventh
	SeventhMinor              // minor seventh
	SeventhDiminished         // diminished seventh
	SeventhOther              // an augmented seventh
)

// String returns the name of the constant, such as "SeventhMinor", for tests
// and logs.
func (s Seventh) String() string {
	names := [...]string{"NoSeventh", "SeventhMajor", "SeventhMinor", "SeventhDiminished", "SeventhOther"}
	if s < 0 || int(s) >= len(names) {
		return fmt.Sprintf("Seventh(%d)", int(s))
	}
	return names[s]
}

// Tension is the view of a ninth, eleventh or thirteenth of a chord.
// Tension9 is the zero value.
type Tension int

// The tension views, in ascending order.
const (
	Tension9       Tension = iota // major ninth
	TensionFlat9                  // minor ninth
	TensionSharp9                 // augmented ninth
	Tension11                     // perfect eleventh
	TensionSharp11                // augmented eleventh
	Tension13                     // major thirteenth
	TensionFlat13                 // minor thirteenth
	// TensionOther is a ninth, eleventh or thirteenth of another quality,
	// such as a diminished eleventh.
	TensionOther
)

// String returns the name of the constant, such as "TensionFlat9", for tests
// and logs.
func (t Tension) String() string {
	names := [...]string{"Tension9", "TensionFlat9", "TensionSharp9", "Tension11", "TensionSharp11", "Tension13", "TensionFlat13", "TensionOther"}
	if t < 0 || int(t) >= len(names) {
		return fmt.Sprintf("Tension(%d)", int(t))
	}
	return names[t]
}

// Triad returns the triad view. With a third, the third and the fifth decide
// it. Without one, a perfect fourth or a major second, with a perfect fifth,
// makes sus4 or sus2 (the fourth or second replaces the third). Anything else,
// including a chord with both a fourth and a second, is TriadOther.
func (c Chord) Triad() Triad {
	fifth, hasFifth := c.has(5)
	if !hasFifth {
		return TriadOther
	}
	if third, ok := c.has(3); ok {
		return triadWithThird(third.quality, fifth.quality)
	}
	fourth, hasFourth := c.has(4)
	second, hasSecond := c.has(2)
	switch {
	case fifth.quality != Perfect || hasFourth == hasSecond:
		return TriadOther
	case hasFourth && fourth.quality == Perfect:
		return TriadSus4
	case hasSecond && second.quality == Major:
		return TriadSus2
	default:
		return TriadOther
	}
}

func triadWithThird(third, fifth Quality) Triad {
	switch {
	case third == Major && fifth == Perfect:
		return TriadMajor
	case third == Minor && fifth == Perfect:
		return TriadMinor
	case third == Minor && fifth == Diminished:
		return TriadDiminished
	case third == Major && fifth == Augmented:
		return TriadAugmented
	default:
		return TriadOther
	}
}

// Seventh returns the seventh view.
func (c Chord) Seventh() Seventh {
	seventh, ok := c.has(7)
	if !ok {
		return NoSeventh
	}
	switch seventh.quality {
	case Major:
		return SeventhMajor
	case Minor:
		return SeventhMinor
	case Diminished:
		return SeventhDiminished
	default:
		return SeventhOther
	}
}

// Tensions returns the views of the ninth, eleventh and thirteenth, in that
// order, for those the chord has. A ninth without a seventh is the added ninth
// (add9): Seventh is NoSeventh and Tensions has Tension9. A ninth with a
// seventh is the ninth of a ninth chord.
func (c Chord) Tensions() []Tension {
	var ts []Tension
	for _, number := range []int{9, 11, 13} {
		if iv, ok := c.has(number); ok {
			ts = append(ts, tensionOf(iv))
		}
	}
	return ts
}

func tensionOf(iv Interval) Tension {
	switch iv {
	case Interval{steps: 8, quality: Major}:
		return Tension9
	case Interval{steps: 8, quality: Minor}:
		return TensionFlat9
	case Interval{steps: 8, quality: Augmented}:
		return TensionSharp9
	case Interval{steps: 10, quality: Perfect}:
		return Tension11
	case Interval{steps: 10, quality: Augmented}:
		return TensionSharp11
	case Interval{steps: 12, quality: Major}:
		return Tension13
	case Interval{steps: 12, quality: Minor}:
		return TensionFlat13
	default:
		return TensionOther
	}
}
