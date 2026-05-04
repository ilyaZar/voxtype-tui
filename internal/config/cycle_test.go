package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestReadCycleFiltersKnownCodes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "language_cycle.toml")
	if err := os.WriteFile(path, []byte(`enabled = ["ru", "xx", "en"]`), 0o644); err != nil {
		t.Fatal(err)
	}

	codes, err := ReadCycle(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"en", "ru"}
	if !reflect.DeepEqual(codes, want) {
		t.Fatalf("codes = %#v, want %#v", codes, want)
	}
}

func TestWriteCycle(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "language_cycle.toml")
	if err := WriteCycle(path, []string{"en", "de"}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "# Languages included in the Ctrl+F12 Voxtype cycle.\nenabled = [\"en\", \"de\"]\n"
	if string(data) != want {
		t.Fatalf("data = %q, want %q", string(data), want)
	}
}
