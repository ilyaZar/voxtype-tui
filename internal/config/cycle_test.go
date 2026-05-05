package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ilyaZar/voxtype-tui/internal/voxtype"
)

var cycleTestLanguages = []voxtype.Language{
	{Code: "en"},
	{Code: "de"},
	{Code: "ru"},
}

func TestReadCycleFiltersKnownCodesAndPreservesOrder(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "language_cycle.toml")
	if err := os.WriteFile(path, []byte(`enabled = ["ru", "xx", "en", "ru"]`), 0o644); err != nil {
		t.Fatal(err)
	}

	codes, err := ReadCycle(path, cycleTestLanguages)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"ru", "en"}
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
	want := "# Languages included in the Ctrl+Pause/Break Voxtype cycle.\nenabled = [\"en\", \"de\"]\n"
	if string(data) != want {
		t.Fatalf("data = %q, want %q", string(data), want)
	}
}

func TestWriteCyclePreservesSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.toml")
	link := filepath.Join(dir, "language_cycle.toml")
	if err := os.WriteFile(target, []byte(`enabled = ["ru"]`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	if err := WriteCycle(link, []string{"en"}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("cycle link was replaced with mode %s", info.Mode())
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `enabled = ["en"]`) {
		t.Fatalf("target was not updated: %s", string(data))
	}
}
