package httpapi

import (
	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/progression"
	"github.com/hassaku63/chord-rambler/backend/internal/query"
	"github.com/hassaku63/chord-rambler/backend/internal/reading"
)

// The one place where the engine's enumerations become the identifiers of
// docs/api/openapi.yaml (lowercase English, words joined by "-"). No switch
// has a default, so the lint (exhaustive) fails when a constant is added; a
// value of none of the constants gives "". The tests walk every value.

func roleID(r progression.Role) string {
	switch r {
	case progression.RoleNone:
		return ""
	case progression.RoleStable:
		return "stable"
	case progression.RoleLink:
		return "link"
	case progression.RolePrepare:
		return "prepare"
	case progression.RoleRise:
		return "rise"
	case progression.RoleResolve:
		return "resolve"
	case progression.RoleClose:
		return "close"
	case progression.RoleSpread:
		return "spread"
	case progression.RoleFloat:
		return "float"
	case progression.RoleWistful:
		return "wistful"
	case progression.RoleLinger:
		return "linger"
	}
	return ""
}

// parseRole returns the role of the identifier, trying every role the slot
// accepts, or false for an identifier of none.
func parseRole(id string) (progression.Role, bool) {
	for r := progression.RoleNone + 1; ; r++ {
		if _, err := (progression.Slot{}).WithRole(r); err != nil {
			return progression.RoleNone, false
		}
		if roleID(r) == id {
			return r, true
		}
	}
}

func basisID(b query.Basis) string {
	switch b {
	case query.Definition:
		return "definition"
	case query.Computation:
		return "computation"
	case query.School:
		return "school"
	case query.Convention:
		return "convention"
	}
	return ""
}

func kindID(k reading.Kind) string {
	switch k {
	case reading.FactsOnly:
		return "facts-only"
	case reading.Diatonic:
		return "diatonic"
	case reading.Borrowed:
		return "borrowed"
	case reading.SecondaryDominant:
		return "secondary-dominant"
	}
	return ""
}

func functionID(f reading.Function) string {
	switch f {
	case reading.NoFunction:
		return "none"
	case reading.Tonic:
		return "tonic"
	case reading.Subdominant:
		return "subdominant"
	case reading.Dominant:
		return "dominant"
	}
	return ""
}

func modeID(m material.Mode) string {
	switch m {
	case material.Ionian:
		return "ionian"
	case material.Dorian:
		return "dorian"
	case material.Phrygian:
		return "phrygian"
	case material.Lydian:
		return "lydian"
	case material.Mixolydian:
		return "mixolydian"
	case material.Aeolian:
		return "aeolian"
	case material.Locrian:
		return "locrian"
	}
	return ""
}

func resolutionID(r reading.Resolution) string {
	switch r {
	case reading.NoResolution:
		return "none"
	case reading.ResolutionExpected:
		return "expected"
	case reading.ResolutionIrregular:
		return "irregular"
	}
	return ""
}

func cadenceID(c reading.Cadence) string {
	switch c {
	case reading.NoCadence:
		return "none"
	case reading.AuthenticCadence:
		return "authentic"
	case reading.HalfCadence:
		return "half"
	case reading.PlagalCadence:
		return "plagal"
	case reading.DeceptiveCadence:
		return "deceptive"
	}
	return ""
}

// roleFitID returns "" for RoleNotJudged: the slot gets no role fit.
func roleFitID(f reading.RoleFit) string {
	switch f {
	case reading.RoleNotJudged:
		return ""
	case reading.RoleFits:
		return "fits"
	case reading.RoleMisfits:
		return "misfits"
	case reading.RoleUndecided:
		return "undecided"
	}
	return ""
}
