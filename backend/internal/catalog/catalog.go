package catalog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
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

// Style is the range of music a source limits what it says to. The zero
// value is never returned by a loader.
type Style int

// The styles a source may be limited to.
const (
	// Classical is Western classical music.
	Classical Style = iota + 1
)

// String returns the name of the constant, for tests and logs.
func (s Style) String() string {
	switch s {
	case Classical:
		return "Classical"
	default:
		return fmt.Sprintf("Style(%d)", int(s))
	}
}

// Source is where the data of a catalogue comes from: the address of the
// document, the version of it that was read, and the style the document
// limits what it says to.
type Source struct {
	URL     string
	Version string
	Style   Style
}

// sourceDoc is the source as written in a catalogue.
type sourceDoc struct {
	URL     string `json:"url"`
	Version string `json:"version"`
	Style   string `json:"style"`
}

var styleWords = map[string]Style{"classical": Classical}

// parseSource checks that the source has an address, a version and a known
// style.
func parseSource(d *sourceDoc) (Source, error) {
	if d == nil || strings.TrimSpace(d.URL) == "" || strings.TrimSpace(d.Version) == "" {
		return Source{}, errors.New("the source needs a url and a version")
	}
	style, ok := styleWords[d.Style]
	if !ok {
		return Source{}, fmt.Errorf("unknown style %q", d.Style)
	}
	return Source{URL: d.URL, Version: d.Version, Style: style}, nil
}

// decodeStrict decodes one JSON document into v. encoding/json matches the
// names of fields without regard to case and keeps the last of repeated
// names, so the names are first checked to be spelt exactly as one of names
// and to appear at most once in each object. A name that does not belong
// where it appears is then refused by the decoder.
func decodeStrict(data []byte, v any, names ...string) error {
	if err := checkNames(data, names); err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return errors.New("data after the document")
	}
	return nil
}

// checkNames walks the tokens of the document and checks the name of every
// member of every object.
func checkNames(data []byte, names []string) error {
	type object struct {
		seen    map[string]bool
		wantKey bool
	}
	var open []*object // innermost last; nil for an array
	dec := json.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		var top *object
		if len(open) > 0 {
			top = open[len(open)-1]
		}
		if name, ok := tok.(string); ok && top != nil && top.wantKey {
			switch {
			case !slices.Contains(names, name):
				return fmt.Errorf("unknown name %q", name)
			case top.seen[name]:
				return fmt.Errorf("name %q appears twice in one object", name)
			}
			top.seen[name] = true
			top.wantKey = false
			continue
		}
		if top != nil {
			top.wantKey = true // after this value, a name or the end
		}
		switch tok {
		case json.Delim('{'):
			open = append(open, &object{seen: map[string]bool{}, wantKey: true})
		case json.Delim('['):
			open = append(open, nil)
		case json.Delim('}'), json.Delim(']'):
			open = open[:len(open)-1]
		}
	}
}
