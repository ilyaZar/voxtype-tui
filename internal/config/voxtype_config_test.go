package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ilyaZar/voxtype-tui/internal/voxtype"
)

func TestCurrentPresetAndWritePreset(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	text := strings.Join([]string{
		"# comment",
		"[whisper]",
		"model = \"base.en\"",
		"language = \"en\"",
		"",
	}, "\n")
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}

	current, ok, err := CurrentPreset(path)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || current.Code != "en" {
		t.Fatalf("current = %#v ok=%v", current, ok)
	}

	de, _ := voxtype.ByCode("de")
	if err := WritePreset(path, de); err != nil {
		t.Fatal(err)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(updated), "model = \"base\"") {
		t.Fatalf("updated config missing base model: %s", string(updated))
	}
	if !strings.Contains(string(updated), "language = \"de\"") {
		t.Fatalf("updated config missing de language: %s", string(updated))
	}
}
