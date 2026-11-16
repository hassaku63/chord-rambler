package catalog_test

import (
	"embed"
	"slices"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/catalog"
)

// catalogueFiles are the JSON files of the package, to find a catalogue that
// has no loader below.
//
//go:embed *.json
var catalogueFiles embed.FS

// loaders loads every catalogue, by the file it embeds. A new catalogue adds
// its file and its loader here.
var loaders = map[string]func() error{
	"function_transitions.json": func() error { _, err := catalog.LoadFunctionTransitions(); return err },
}

// TestEveryCatalogueLoads loads and checks every catalogue embedded in the
// binary (ADR-0008, 6).
func TestEveryCatalogueLoads(t *testing.T) {
	entries, err := catalogueFiles.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	var files []string
	for _, e := range entries {
		files = append(files, e.Name())
	}
	var loaded []string
	for name := range loaders {
		loaded = append(loaded, name)
	}
	slices.Sort(loaded)
	if !slices.Equal(files, loaded) {
		t.Errorf("JSON files = %v; loaders = %v; want the same", files, loaded)
	}
	for _, name := range loaded {
		t.Run(name, func(t *testing.T) {
			if err := loaders[name](); err != nil {
				t.Errorf("load %s: %v", name, err)
			}
		})
	}
}
