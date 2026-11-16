package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"reflect"
	"strconv"
	"strings"

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

// maxBodyBytes is the largest body the API reads. Append copies the progression
// for every slot, so the work grows with the square of the slots; a bound on
// the bytes bounds it.
const maxBodyBytes = 64 << 10

// requestError is a failure of the request, with the status it is answered by.
// Its detail says where and what, never how the server reads it.
type requestError struct {
	status int
	detail string
}

func (e *requestError) Error() string { return e.detail }

func badRequest(format string, args ...any) error {
	return &requestError{status: http.StatusBadRequest, detail: fmt.Sprintf(format, args...)}
}

// decodeBody is the one place a request body is read into a value: it checks
// the media type (415), cuts the size (413), refuses a field the contract does
// not have (400) and a value of the wrong type (400), and wants one JSON value.
func decodeBody(w http.ResponseWriter, r *http.Request, dst any) error {
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != "application/json" {
		return &requestError{status: http.StatusUnsupportedMediaType, detail: "the Content-Type must be application/json"}
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return describeDecodeError(err)
	}
	if err := dec.Decode(new(json.RawMessage)); !errors.Is(err, io.EOF) {
		if tooLarge := (*http.MaxBytesError)(nil); errors.As(err, &tooLarge) {
			return describeDecodeError(err)
		}
		return badRequest("the body is not one JSON value: it goes on after the progression")
	}
	return nil
}

func describeDecodeError(err error) error {
	var tooLarge *http.MaxBytesError
	var typeErr *json.UnmarshalTypeError
	switch {
	case errors.As(err, &tooLarge):
		return &requestError{status: http.StatusRequestEntityTooLarge,
			detail: fmt.Sprintf("the body is larger than %d bytes", tooLarge.Limit)}
	case errors.As(err, &typeErr):
		return badRequest("%s: expected %s", fieldPath(typeErr.Field), jsonTypeName(typeErr.Type))
	}
	if name, ok := strings.CutPrefix(err.Error(), "json: unknown field "); ok {
		return badRequest("unknown field %s", name)
	}
	return badRequest("the body is not a progression in JSON")
}

// fieldPath writes the path of a field as the contract's reader sees it:
// slots.0.beats becomes slots[0].beats.
func fieldPath(field string) string {
	if field == "" {
		return "body"
	}
	var b strings.Builder
	for i, part := range strings.Split(field, ".") {
		if _, err := strconv.Atoi(part); err == nil {
			b.WriteString("[" + part + "]")
			continue
		}
		if i > 0 {
			b.WriteByte('.')
		}
		b.WriteString(part)
	}
	return b.String()
}

// jsonTypeName names a Go type by the JSON type that fits it.
func jsonTypeName(t reflect.Type) string {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice, reflect.Array:
		return "array"
	default:
		return "object"
	}
}

// decodeProgression reads the body with the engine's constructors, which check
// it. Every error it returns is a *requestError.
func decodeProgression(w http.ResponseWriter, r *http.Request) (progression.Progression, error) {
	var body progressionBody
	if err := decodeBody(w, r, &body); err != nil {
		return progression.Progression{}, err
	}
	tonic, err := notation.ParseNoteName(body.Key)
	if err != nil {
		return progression.Progression{}, badRequest("key: %v", err)
	}
	key, err := material.NewScale(tonic, material.Ionian)
	if err != nil {
		return progression.Progression{}, badRequest("key %q: %v", body.Key, err)
	}
	p := progression.New(key, progression.Meter{})
	for i, sb := range body.Slots {
		s, err := decodeSlot(sb)
		if err != nil {
			return progression.Progression{}, badRequest("slot %d: %v", i, err)
		}
		if p, err = p.Append(s); err != nil {
			return progression.Progression{}, badRequest("slot %d: %v", i, err)
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
		p, err := decodeProgression(w, r)
		if err != nil {
			a.writeRequestError(w, err)
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

// writeRequestError answers a failed decode with the status its error carries;
// an error that is not a *requestError is a 400.
func (a *api) writeRequestError(w http.ResponseWriter, err error) {
	var re *requestError
	if !errors.As(err, &re) {
		re = &requestError{status: http.StatusBadRequest, detail: err.Error()}
	}
	a.writeProblem(w, re.status, re.detail)
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
