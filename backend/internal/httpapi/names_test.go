package httpapi

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/catalog"
	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/progression"
	"github.com/hassaku63/chord-rambler/backend/internal/query"
	"github.com/hassaku63/chord-rambler/backend/internal/reading"
)

// valuesOf walks the values of an enumeration from first upward while its
// String is a constant's name, not the "Type(n)" of a value of none, so that
// every constant is visited without knowing the last one.
func valuesOf[T ~int](first T) []T {
	var out []T
	for v := first; ; v++ {
		if strings.HasSuffix(fmt.Sprint(v), fmt.Sprintf("(%d)", int(v))) {
			return out
		}
		out = append(out, v)
	}
}

// idsOf maps every value and checks that each one past the last maps to "".
func idsOf[T ~int](t *testing.T, values []T, id func(T) string) []string {
	t.Helper()
	out := make([]string, 0, len(values))
	for _, v := range values {
		out = append(out, id(v))
	}
	past := values[len(values)-1] + 1
	if got := id(past); got != "" {
		t.Errorf("%v maps to %q, want \"\"", past, got)
	}
	return out
}

// TestEveryValueHasAnIdentifier walks every value of each enumeration the API
// writes and checks the identifiers, and that docs/api/openapi.yaml lists the
// same ones ("" is a value the API leaves out).
func TestEveryValueHasAnIdentifier(t *testing.T) {
	contract, err := os.ReadFile(openAPIPath)
	if err != nil {
		t.Fatalf("reading the contract: %v", err)
	}
	tests := []struct {
		name string
		got  func(t *testing.T) []string
		want []string
	}{
		{"Role", func(t *testing.T) []string {
			return idsOf(t, valuesOf(progression.RoleNone), roleID)
		}, []string{"", "stable", "link", "prepare", "rise", "resolve", "close", "spread", "float", "wistful", "linger"}},
		{"Basis", func(t *testing.T) []string {
			return idsOf(t, valuesOf(query.Definition), basisID)
		}, []string{"definition", "computation", "school", "convention"}},
		{"Kind", func(t *testing.T) []string {
			return idsOf(t, valuesOf(reading.FactsOnly), kindID)
		}, []string{"facts-only", "diatonic", "borrowed", "secondary-dominant"}},
		{"Function", func(t *testing.T) []string {
			return idsOf(t, valuesOf(reading.NoFunction), functionID)
		}, []string{"none", "tonic", "subdominant", "dominant"}},
		{"Mode", func(t *testing.T) []string {
			return idsOf(t, valuesOf(material.Ionian), modeID)
		}, []string{"ionian", "dorian", "phrygian", "lydian", "mixolydian", "aeolian", "locrian"}},
		{"Resolution", func(t *testing.T) []string {
			return idsOf(t, valuesOf(reading.NoResolution), resolutionID)
		}, []string{"none", "expected", "irregular"}},
		{"Cadence", func(t *testing.T) []string {
			return idsOf(t, valuesOf(reading.NoCadence), cadenceID)
		}, []string{"none", "authentic", "half", "plagal", "deceptive"}},
		{"RoleFit", func(t *testing.T) []string {
			return idsOf(t, valuesOf(reading.RoleNotJudged), roleFitID)
		}, []string{"", "fits", "misfits", "undecided"}},
		{"RoleCondition", func(t *testing.T) []string {
			return idsOf(t, valuesOf(reading.ConditionShape), conditionID)
		}, []string{"shape", "function", "previous-v", "next-dominant", "cadence", "another-root", "expected-resolution"}},
		{"Preference", func(t *testing.T) []string {
			return idsOf(t, valuesOf(catalog.Common), preferenceID)
		}, []string{"common", "less-common", "unstated"}},
		{"Style", func(t *testing.T) []string {
			return idsOf(t, valuesOf(catalog.Classical), styleID)
		}, []string{"classical"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.got(t); !slices.Equal(got, tt.want) {
				t.Errorf("identifiers = %q, want %q", got, tt.want)
			}
			enum := "enum: [" + strings.Join(slices.DeleteFunc(slices.Clone(tt.want), func(s string) bool { return s == "" }), ", ") + "]"
			if !strings.Contains(string(contract), enum) {
				t.Errorf("%s has no %q", openAPIPath, enum)
			}
		})
	}
}

// TestParseRoleReadsEveryIdentifier reads back every role's identifier.
func TestParseRoleReadsEveryIdentifier(t *testing.T) {
	for _, r := range valuesOf(progression.RoleStable) {
		got, ok := parseRole(roleID(r))
		if !ok || got != r {
			t.Errorf("parseRole(%q) = %v, %v, want %v, true", roleID(r), got, ok, r)
		}
	}
	for _, id := range []string{"", "none", "Stable", "RoleStable", "安定"} {
		if r, ok := parseRole(id); ok {
			t.Errorf("parseRole(%q) = %v, true, want false", id, r)
		}
	}
}
