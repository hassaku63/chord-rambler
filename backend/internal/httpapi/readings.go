package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/notation"
	"github.com/hassaku63/chord-rambler/backend/internal/progression"
	"github.com/hassaku63/chord-rambler/backend/internal/query"
	"github.com/hassaku63/chord-rambler/backend/internal/reading"
	"github.com/hassaku63/chord-rambler/backend/internal/relation"
)

// progressionBody is the Progression of the contract.
type progressionBody struct {
	Key   string     `json:"key"`
	Slots []slotBody `json:"slots"`
}

// slotBody is the Slot of the contract; a field left out is nil.
type slotBody struct {
	Chord   *string `json:"chord"`
	Content *string `json:"content"`
	Beats   int     `json:"beats"`
	Role    *string `json:"role"`
}

// decodeProgression reads the body with the engine's constructors, which check
// it. Every error it returns is the client's and says what could not be read.
func decodeProgression(r *http.Request) (progression.Progression, error) {
	var body progressionBody
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&body); err != nil {
		return progression.Progression{}, fmt.Errorf("the body is not a progression in JSON: %w", err)
	}
	if dec.More() {
		return progression.Progression{}, errors.New("the body is not one JSON value: it goes on after the progression")
	}
	tonic, err := notation.ParseNoteName(body.Key)
	if err != nil {
		return progression.Progression{}, fmt.Errorf("key: %w", err)
	}
	key, err := material.NewScale(tonic, material.Ionian)
	if err != nil {
		return progression.Progression{}, fmt.Errorf("key %q: %w", body.Key, err)
	}
	p := progression.New(key, progression.Meter{})
	for i, sb := range body.Slots {
		s, err := decodeSlot(sb)
		if err != nil {
			return progression.Progression{}, fmt.Errorf("slot %d: %w", i, err)
		}
		if p, err = p.Append(s); err != nil {
			return progression.Progression{}, fmt.Errorf("slot %d: %w", i, err)
		}
	}
	return p, nil
}

func decodeSlot(sb slotBody) (progression.Slot, error) {
	var content progression.Content
	switch {
	case (sb.Chord == nil) == (sb.Content == nil):
		return progression.Slot{}, errors.New("give exactly one of chord and content")
	case sb.Chord != nil:
		c, err := notation.ParseChord(*sb.Chord)
		if err != nil {
			return progression.Slot{}, err
		}
		content = progression.ChordOf(c)
	case *sb.Content == "none":
		content = progression.NoChord()
	case *sb.Content == "unchosen":
		content = progression.Unchosen()
	default:
		return progression.Slot{}, fmt.Errorf("content %q is neither none nor unchosen", *sb.Content)
	}
	s, err := progression.NewSlot(content, sb.Beats)
	if err != nil {
		return progression.Slot{}, err
	}
	if sb.Role == nil {
		return s, nil
	}
	role, ok := parseRole(*sb.Role)
	if !ok {
		return progression.Slot{}, fmt.Errorf("role %q: %w", *sb.Role, progression.ErrRole)
	}
	return s.WithRole(role)
}

// The labels of the contract's SlotReading.
type (
	progressionReading struct {
		Slots []slotReading `json:"slots"`
	}
	slotReading struct {
		Chord           *chordNameLabel       `json:"chord,omitempty"`
		Roman           *romanLabel           `json:"roman,omitempty"`
		Reading         *chordReadingLabel    `json:"reading,omitempty"`
		Resolution      *resolutionLabel      `json:"resolution,omitempty"`
		Cadence         *cadenceLabel         `json:"cadence,omitempty"`
		TensionCautions *tensionCautionsLabel `json:"tensionCautions,omitempty"`
		RoleFit         *roleFitLabel         `json:"roleFit,omitempty"`
	}
	chordNameLabel struct {
		Basis string `json:"basis"`
		Name  string `json:"name"`
	}
	romanLabel struct {
		Basis   string `json:"basis"`
		Numeral string `json:"numeral"`
	}
	chordReadingLabel struct {
		Basis             string   `json:"basis"`
		Kind              string   `json:"kind"`
		Function          string   `json:"function"`
		Modes             []string `json:"modes"`
		FromParallelMinor bool     `json:"fromParallelMinor"`
		Target            int      `json:"target,omitempty"`
	}
	resolutionLabel struct {
		Basis      string `json:"basis"`
		Resolution string `json:"resolution"`
	}
	cadenceLabel struct {
		Basis   string `json:"basis"`
		Cadence string `json:"cadence"`
	}
	tensionCautionsLabel struct {
		Basis    string           `json:"basis"`
		Cautions []tensionCaution `json:"cautions"`
	}
	tensionCaution struct {
		Tension   string `json:"tension"`
		ChordTone string `json:"chordTone"`
	}
	roleFitLabel struct {
		Basis string `json:"basis"`
		Fit   string `json:"fit"`
	}
)

func (a *api) readings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := decodeProgression(r)
		if err != nil {
			a.writeProblem(w, http.StatusBadRequest, err.Error())
			return
		}
		out, err := readProgression(p)
		if err != nil {
			a.logger.Error("reading the progression failed", "error", err)
			a.writeProblem(w, http.StatusInternalServerError, "")
			return
		}
		a.writeJSON(w, http.StatusOK, "application/json", out)
	}
}

// readProgression calls the engine and writes its answers as labels.
func readProgression(p progression.Progression) (progressionReading, error) {
	read, err := reading.ReadProgression(p)
	if err != nil {
		return progressionReading{}, err
	}
	fits, err := reading.ReadRoleFits(p)
	if err != nil {
		return progressionReading{}, err
	}
	slots := p.Slots()
	out := progressionReading{Slots: make([]slotReading, len(slots))}
	for i, s := range slots {
		if c, ok := s.Content().Chord(); ok {
			if out.Slots[i], err = chordLabels(c, p.Key(), read[i]); err != nil {
				return progressionReading{}, fmt.Errorf("slot %d: %w", i, err)
			}
		}
		if id := roleFitID(fits[i]); id != "" {
			out.Slots[i].RoleFit = &roleFitLabel{Basis: basisID(query.School), Fit: id}
		}
	}
	return out, nil
}

func chordLabels(c material.Chord, key material.Scale, r reading.SlotReading) (slotReading, error) {
	name, err := notation.FormatChord(c)
	if err != nil {
		return slotReading{}, err
	}
	cautions, err := reading.TensionCautions(c, key)
	if err != nil {
		return slotReading{}, err
	}
	roman, err := romanOf(c, key, r.Chord.Main)
	if err != nil {
		return slotReading{}, err
	}
	main := r.Chord.Main
	modes := make([]string, 0, len(main.Modes))
	for _, m := range main.Modes {
		modes = append(modes, modeID(m))
	}
	tc := make([]tensionCaution, 0, len(cautions))
	for _, tcn := range cautions {
		tc = append(tc, tensionCaution{Tension: tcn.Tension.String(), ChordTone: tcn.ChordTone.String()})
	}
	school := basisID(query.School)
	return slotReading{
		Chord: &chordNameLabel{Basis: basisID(query.Definition), Name: name},
		Roman: roman,
		Reading: &chordReadingLabel{
			Basis: school, Kind: kindID(main.Kind), Function: functionID(main.Function),
			Modes: modes, FromParallelMinor: main.FromParallelMinor(), Target: main.Target,
		},
		Resolution:      &resolutionLabel{Basis: school, Resolution: resolutionID(r.Resolution)},
		Cadence:         &cadenceLabel{Basis: school, Cadence: cadenceID(r.Cadence)},
		TensionCautions: &tensionCautionsLabel{Basis: school, Cautions: tc},
	}, nil
}

// romanOf writes the chord as a Roman numeral, without its bass, which Roman
// numerals do not write (docs/architecture/theory-conventions.md, 6). A
// secondary dominant is written V/x, as its reading says (#83), and the label
// rests on the school; any other chord is written at its position in the key,
// a computation. A root with no position gives no label (nil).
func romanOf(c material.Chord, key material.Scale, main reading.Reading) (*romanLabel, error) {
	shape := material.ShapeOf(c)
	if _, ok := c.Bass(); ok {
		noBass, err := material.NewChord(c.Root(), c.Intervals()[1:]...)
		if err != nil {
			return nil, err
		}
		shape = material.ShapeOf(noBass)
	}
	var pos material.Position
	basis := query.Computation
	if main.Kind == reading.SecondaryDominant {
		five, err := material.NewPosition(5, material.Natural, shape)
		if err != nil {
			return nil, err
		}
		if pos, err = five.WithTarget(main.Target); err != nil {
			return nil, err
		}
		basis = query.School
	} else {
		at, err := relation.PositionOf(c, key)
		if errors.Is(err, relation.ErrNoPosition) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		if pos, err = material.NewPosition(at.Degree(), at.Alteration(), shape); err != nil {
			return nil, err
		}
	}
	numeral, err := notation.FormatRoman(pos)
	if err != nil {
		return nil, err
	}
	return &romanLabel{Basis: basisID(basis), Numeral: numeral}, nil
}
