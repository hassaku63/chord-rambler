package catalog

import (
	"bytes"
	_ "embed" // the catalogue is embedded in the binary (ADR-0008, 6)
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hassaku63/chord-rambler/backend/internal/reading"
)

// ErrInvalid is wrapped by every error from loading a catalogue that is not
// well formed.
var ErrInvalid = errors.New("invalid catalogue")

// Grounds is the kind of knowledge an entry of a catalogue rests on
// (docs/architecture/explanations.md, "根拠の種類ごとの言い方"). The zero value
// is never returned by a loader.
type Grounds int

// The kinds of grounds a catalogue may hold.
const (
	// Convention is a preference of a style: said as a frequency, never as
	// right or wrong.
	Convention Grounds = iota + 1
)

// String returns the name of the constant, for tests and logs.
func (g Grounds) String() string {
	switch g {
	case Convention:
		return "Convention"
	default:
		return fmt.Sprintf("Grounds(%d)", int(g))
	}
}

// Preference is how often the source says one function moves to another. The
// zero value is never returned by a loader.
type Preference int

// The preferences, named after what the source says.
const (
	// Common is a move the source says harmonies tend to make.
	Common Preference = iota + 1
	// LessCommon is a move the source says tends to be less common.
	LessCommon
	// Unstated is a move the source says nothing about how often it is
	// made. It is not a frequency, so no label of convention is shown for it.
	Unstated
)

// String returns the name of the constant, for tests and logs.
func (p Preference) String() string {
	switch p {
	case Common:
		return "Common"
	case LessCommon:
		return "LessCommon"
	case Unstated:
		return "Unstated"
	default:
		return fmt.Sprintf("Preference(%d)", int(p))
	}
}

// Source is where the data of a catalogue comes from: the address of the
// document and the version of it that was read.
type Source struct {
	URL     string
	Version string
}

// Transition is one entry of the catalogue of function transitions: how often
// a chord of the function From is followed by one of the function To.
type Transition struct {
	From, To   reading.Function
	Preference Preference
	Grounds    Grounds
}

// FunctionTransitions is the catalogue of function transitions: one entry
// for every ordered pair of tonic, subdominant and dominant. Only the loaders
// make one, so every value holds all nine pairs.
type FunctionTransitions struct {
	source  Source
	entries [3][3]Transition // by From-1, then To-1
}

// Source returns where the catalogue comes from.
func (c FunctionTransitions) Source() Source { return c.source }

// Lookup returns the entry from one function to the next. It reports false
// when either is not tonic, subdominant or dominant.
func (c FunctionTransitions) Lookup(from, to reading.Function) (Transition, bool) {
	i, ok := functionIndex(from)
	j, ok2 := functionIndex(to)
	if !ok || !ok2 {
		return Transition{}, false
	}
	return c.entries[i][j], true
}

//go:embed function_transitions.json
var functionTransitionsJSON []byte

// LoadFunctionTransitions parses and checks the catalogue embedded in the
// binary.
func LoadFunctionTransitions() (FunctionTransitions, error) {
	return ParseFunctionTransitions(functionTransitionsJSON)
}

// The words of the JSON document.
var (
	functionWords = map[string]reading.Function{
		"T": reading.Tonic, "S": reading.Subdominant, "D": reading.Dominant,
	}
	preferenceWords = map[string]Preference{
		"common": Common, "less_common": LessCommon, "unstated": Unstated,
	}
	groundsWords = map[string]Grounds{"convention": Convention}
)

type transitionsDoc struct {
	Source *struct {
		URL     string `json:"url"`
		Version string `json:"version"`
	} `json:"source"`
	Transitions []struct {
		From       string `json:"from"`
		To         string `json:"to"`
		Preference string `json:"preference"`
		Grounds    string `json:"grounds"`
	} `json:"transitions"`
}

// ParseFunctionTransitions parses a catalogue of function transitions and
// checks it: the source has an address and a version, every value is known,
// and each of the nine pairs of functions appears exactly once. Fields that
// are not part of the catalogue are errors too. Every error wraps ErrInvalid.
func ParseFunctionTransitions(data []byte) (FunctionTransitions, error) {
	out, err := parseFunctionTransitions(data)
	if err != nil {
		return FunctionTransitions{}, fmt.Errorf("%w: function transitions: %w", ErrInvalid, err)
	}
	return out, nil
}

func parseFunctionTransitions(data []byte) (FunctionTransitions, error) {
	var doc transitionsDoc
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&doc); err != nil {
		return FunctionTransitions{}, err
	}
	if dec.More() {
		return FunctionTransitions{}, errors.New("data after the document")
	}
	if doc.Source == nil || strings.TrimSpace(doc.Source.URL) == "" || strings.TrimSpace(doc.Source.Version) == "" {
		return FunctionTransitions{}, errors.New("the source needs a url and a version")
	}
	out := FunctionTransitions{source: Source{URL: doc.Source.URL, Version: doc.Source.Version}}
	var seen [3][3]bool
	for n, e := range doc.Transitions {
		t := Transition{
			From:       functionWords[e.From],
			To:         functionWords[e.To],
			Preference: preferenceWords[e.Preference],
			Grounds:    groundsWords[e.Grounds],
		}
		i, ok := functionIndex(t.From)
		j, ok2 := functionIndex(t.To)
		switch {
		case !ok || !ok2:
			return FunctionTransitions{}, fmt.Errorf("entry %d: unknown function %q → %q", n, e.From, e.To)
		case t.Preference == 0:
			return FunctionTransitions{}, fmt.Errorf("entry %d: unknown preference %q", n, e.Preference)
		case t.Grounds == 0:
			return FunctionTransitions{}, fmt.Errorf("entry %d: unknown grounds %q", n, e.Grounds)
		case seen[i][j]:
			return FunctionTransitions{}, fmt.Errorf("entry %d: %s → %s appears twice", n, e.From, e.To)
		}
		seen[i][j] = true
		out.entries[i][j] = t
	}
	if missing := missingPairs(seen); len(missing) > 0 {
		return FunctionTransitions{}, fmt.Errorf("missing %s", strings.Join(missing, ", "))
	}
	return out, nil
}

// functions are tonic, subdominant and dominant, in the order of the table of
// entries.
var functions = [3]reading.Function{reading.Tonic, reading.Subdominant, reading.Dominant}

// missingPairs names the pairs of functions not seen.
func missingPairs(seen [3][3]bool) []string {
	var out []string
	for i, from := range functions {
		for j, to := range functions {
			if !seen[i][j] {
				out = append(out, fmt.Sprintf("%v → %v", from, to))
			}
		}
	}
	return out
}

// functionIndex returns the index of tonic, subdominant and dominant in the
// table of entries, and false for any other function.
func functionIndex(f reading.Function) (int, bool) {
	switch f {
	case reading.Tonic:
		return 0, true
	case reading.Subdominant:
		return 1, true
	case reading.Dominant:
		return 2, true
	case reading.NoFunction:
		return 0, false
	default:
		return 0, false
	}
}
