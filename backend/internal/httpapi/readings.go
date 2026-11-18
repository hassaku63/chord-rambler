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

// maxSlots is the most slots a progression of a body may have: 64 bars of 4/4
// cut into single beats. The bytes alone allow some 2,000 short slots, and
// the work on them grows with the square of the number.
const maxSlots = 256

// requestError is a failure of the request, with the status it is answered by.
// Its detail says where and what, never how the server reads it.
type requestError struct {
	status int
	detail string
	// slot is the slot the failure is in, when it is in one.
	slot *int
	// cause is the engine's error behind the detail. Only the log gets its text.
	cause error
}

func (e *requestError) Error() string { return e.detail }

func badRequest(format string, args ...any) error {
	return &requestError{status: http.StatusBadRequest, detail: fmt.Sprintf(format, args...)}
}

// refusal is what a user is told when the engine refuses a request with the
// error: one sentence, with no name of the engine in it.
type refusal struct {
	err  error
	text string
}

// refusals is the one table of the engine's sentinel errors that are a 400.
// The engine's own text, which names Go identifiers and functions, goes to the
// log. An error that is not here is the server's (500).
var refusals = []refusal{
	{progression.ErrIndex, "the slot is outside the progression"},
	{progression.ErrLength, "a slot must be at least one beat long"},
	{progression.ErrCrossesPhrase, "a slot must not cross the boundary of a phrase"},
	{material.ErrAlterOutOfRange, "a note would need more than a double sharp or a double flat"},
	{query.ErrNoTriadQuality, "variations are made only for a chord whose triad is major, minor or diminished"},
}

// refusalOf returns the sentence of the first refusal the error is.
func refusalOf(err error) (string, bool) {
	for _, r := range refusals {
		if errors.Is(err, r.err) {
			return r.text, true
		}
	}
	return "", false
}

// engineRefused is the 400 for an error of the engine: the sentence of the
// table when it has one, else the error's own text. The prefix says where.
func engineRefused(where string, err error) error {
	return &requestError{status: http.StatusBadRequest, detail: where + ": " + refusalText(err), cause: err}
}

// slotRefused is engineRefused for an error in the slot: the detail is
// prefixed "slot N" and the problem carries N.
func slotRefused(slot int, err error) error {
	return &requestError{status: http.StatusBadRequest,
		detail: fmt.Sprintf("slot %d: %s", slot, refusalText(err)), slot: &slot, cause: err}
}

func refusalText(err error) string {
	if t, ok := refusalOf(err); ok {
		return t
	}
	return err.Error()
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
		return badRequest("the body is not one JSON value: it goes on after the first")
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
	return badRequest("the body is not valid JSON")
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
	return body.progression()
}

// parseKey reads the tonic of a major key. Every error it returns is a
// *requestError.
func parseKey(name string) (material.Scale, error) {
	tonic, err := notation.ParseNoteName(name)
	if err != nil {
		return material.Scale{}, badRequest("key: %v", err)
	}
	key, err := material.NewScale(tonic, material.Ionian)
	if err != nil {
		return material.Scale{}, engineRefused(fmt.Sprintf("key %q", name), err)
	}
	return key, nil
}

// progression builds the progression of the body with the engine's
// constructors, which check it. Every error it returns is a *requestError.
func (body progressionBody) progression() (progression.Progression, error) {
	if len(body.Slots) > maxSlots {
		return progression.Progression{}, badRequest("a progression has at most %d slots", maxSlots)
	}
	key, err := parseKey(body.Key)
	if err != nil {
		return progression.Progression{}, err
	}
	p := progression.New(key, progression.Meter{})
	for i, sb := range body.Slots {
		s, err := decodeSlot(sb)
		if err != nil {
			return progression.Progression{}, slotRefused(i, err)
		}
		if p, err = p.Append(s); err != nil {
			return progression.Progression{}, slotRefused(i, err)
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

// The labels of the contract's SlotReading. Content tells the two shapes
// apart: a slot with a chord (chord) always has the chord's labels; a slot
// with no chord (none, unchosen) has only the role fit, if any.
type (
	progressionReading struct {
		Slots []slotReading `json:"slots"`
	}
	slotReading struct {
		Content         string                `json:"content"`
		Chord           *chordNameLabel       `json:"chord,omitempty"`
		Roman           *romanLabel           `json:"roman,omitempty"`
		Reading         *chordReadingLabel    `json:"reading,omitempty"`
		Resolution      *resolutionLabel      `json:"resolution,omitempty"`
		Cadence         *cadenceLabel         `json:"cadence,omitempty"`
		TensionCautions *tensionCautionsLabel `json:"tensionCautions,omitempty"`
		OutOfKey        *notesLabel           `json:"outOfKey,omitempty"`
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
	if re.cause != nil {
		a.logger.Info("the engine refused the request", "error", re.cause)
	}
	a.writeProblemIn(w, re.status, re.detail, re.slot)
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
		out.Slots[i].Content = contentID(s.Content())
		if id := roleFitID(fits[i]); id != "" {
			out.Slots[i].RoleFit = &roleFitLabel{Basis: basisID(query.School), Fit: id}
		}
	}
	return out, nil
}

func chordLabels(c material.Chord, key material.Scale, r reading.SlotReading) (slotReading, error) {
	name, err := chordNameOf(c)
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
	school := query.School
	return slotReading{
		Chord:           name,
		Roman:           roman,
		Reading:         chordReadingOf(school, r.Chord.Main),
		Resolution:      &resolutionLabel{Basis: basisID(school), Resolution: resolutionID(r.Resolution)},
		Cadence:         &cadenceLabel{Basis: basisID(school), Cadence: cadenceID(r.Cadence)},
		TensionCautions: tensionCautionsOf(school, cautions),
		OutOfKey:        outOfKeyOf(query.OutOfKey(c, key)),
	}, nil
}

// outOfKeyOf writes the notes outside the key, or no label (nil) when there
// are none.
func outOfKeyOf(label query.OutOfKeyLabel) *notesLabel {
	if len(label.Notes) == 0 {
		return nil
	}
	notes := notesOf(label.Basis(), label.Notes)
	return &notes
}

// chordNameOf writes the chord with its spelling kept; the name rests on the
// definition.
func chordNameOf(c material.Chord) (*chordNameLabel, error) {
	name, err := notation.FormatChord(c)
	if err != nil {
		return nil, err
	}
	return &chordNameLabel{Basis: basisID(query.Definition), Name: name}, nil
}

// chordReadingOf writes the main reading of a chord.
func chordReadingOf(basis query.Basis, main reading.Reading) *chordReadingLabel {
	modes := make([]string, 0, len(main.Modes))
	for _, m := range main.Modes {
		modes = append(modes, modeID(m))
	}
	return &chordReadingLabel{
		Basis: basisID(basis), Kind: kindID(main.Kind), Function: functionID(main.Function),
		Modes: modes, FromParallelMinor: main.FromParallelMinor(), Target: main.Target,
	}
}

// tensionCautionsOf writes the tensions that tend to sound muddy, always as a
// list.
func tensionCautionsOf(basis query.Basis, cautions []reading.TensionCaution) *tensionCautionsLabel {
	tc := make([]tensionCaution, 0, len(cautions))
	for _, c := range cautions {
		tc = append(tc, tensionCaution{Tension: c.Tension.String(), ChordTone: c.ChordTone.String()})
	}
	return &tensionCautionsLabel{Basis: basisID(basis), Cautions: tc}
}

// romanOf writes the chord as a Roman numeral at the position the reading
// gives it (reading.RomanPosition). A chord with no position gives no label
// (nil).
func romanOf(c material.Chord, key material.Scale, main reading.Reading) (*romanLabel, error) {
	pos, ok, err := reading.RomanPosition(c, key, main)
	if err != nil || !ok {
		return nil, err
	}
	return romanAt(pos, main)
}

// romanAt writes the position as a Roman numeral. A secondary dominant's
// label rests on the school, as its reading says (#83); any other chord's on
// a computation.
func romanAt(pos material.Position, main reading.Reading) (*romanLabel, error) {
	numeral, err := notation.FormatRoman(pos)
	if err != nil {
		return nil, err
	}
	basis := query.Computation
	if main.Kind == reading.SecondaryDominant {
		basis = query.School
	}
	return &romanLabel{Basis: basisID(basis), Numeral: numeral}, nil
}
