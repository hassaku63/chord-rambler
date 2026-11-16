package reading

import (
	"errors"
	"fmt"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/relation"
)

// RomanPosition returns the position at which the chord is written as a Roman
// numeral in the key, given its main reading. The position has no bass: Roman
// numerals write neither inversions nor basses
// (docs/architecture/theory-conventions.md, 6). A secondary dominant is at V
// with the degree it points to (V7/V), as its reading says (#83); any other
// chord is at relation.PositionOf. A chord with no position gives false.
func RomanPosition(c material.Chord, key material.Scale, main Reading) (material.Position, bool, error) {
	shape := material.ShapeOf(c)
	if _, ok := c.Bass(); ok {
		noBass, err := material.NewChord(c.Root(), c.Intervals()[1:]...)
		if err != nil {
			return material.Position{}, false, fmt.Errorf("roman position: %w", err)
		}
		shape = material.ShapeOf(noBass)
	}
	if main.Kind == SecondaryDominant {
		five, err := material.NewPosition(5, material.Natural, shape)
		if err != nil {
			return material.Position{}, false, fmt.Errorf("roman position: %w", err)
		}
		pos, err := five.WithTarget(main.Target)
		if err != nil {
			return material.Position{}, false, fmt.Errorf("roman position: %w", err)
		}
		return pos, true, nil
	}
	at, err := relation.PositionOf(c, key)
	if errors.Is(err, relation.ErrNoPosition) {
		return material.Position{}, false, nil
	}
	if err != nil {
		return material.Position{}, false, fmt.Errorf("roman position: %w", err)
	}
	pos, err := material.NewPosition(at.Degree(), at.Alteration(), shape)
	if err != nil {
		return material.Position{}, false, fmt.Errorf("roman position: %w", err)
	}
	return pos, true, nil
}
