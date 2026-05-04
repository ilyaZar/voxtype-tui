package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ilyaZar/voxtype-tui/internal/voxtype"
)

var configTestLanguages = []voxtype.Language{
	{Code: "en", Name: "English", Label: "EN", Model: "base.en", Language: "en"},
	{Code: "de", Name: "German", Label: "DE", Model: "base", Language: "de"},
	{Code: "ru", Name: "Russian", Label: "RU", Model: "base", Language: "ru"},
}

func TestCurrentPresetAndWritePreset(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	text := strings.Join([]string{
		"# comment",
		"[whisper] # inline comments are valid TOML",
		"  model = 'base.en'",
		"  language = 'en'",
		"",
		"[output] # another section",
		"model = \"base\"",
		"language = \"ru\"",
		"",
	}, "\n")
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}

	current, ok, err := CurrentPreset(path, configTestLanguages)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || current.Code != "en" {
		t.Fatalf("current = %#v ok=%v", current, ok)
	}

	de, _ := voxtype.ByCode("de", configTestLanguages)
	if err := WritePreset(path, de); err != nil {
		t.Fatal(err)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(updated), "  model = \"base\"") {
		t.Fatalf("updated config missing base model: %s", string(updated))
	}
	if !strings.Contains(string(updated), "  language = \"de\"") {
		t.Fatalf("updated config missing de language: %s", string(updated))
	}
	if !strings.Contains(string(updated), "language = \"ru\"") {
		t.Fatalf("updated config changed another section: %s", string(updated))
	}
}

func TestReadLanguagesFromVoxtypeTUISection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	text := strings.Join([]string{
		"[[voxtype_tui.language]]",
		"code = 'no'",
		"name = 'Norwegian'",
		"label = 'NO'",
		"model = 'base'",
		"language = 'no'",
		"",
	}, "\n")
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}

	languages, err := ReadLanguages(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(languages) != 1 || languages[0].Code != "no" || languages[0].Name != "Norwegian" {
		t.Fatalf("languages = %#v", languages)
	}
}

func TestReadLanguagesErrorsWithoutConfiguredLanguages(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	link := filepath.Join(dir, "linked-config.toml")
	if err := os.WriteFile(path, []byte("[whisper]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(path, link); err != nil {
		t.Fatal(err)
	}

	_, err := ReadLanguages(link)
	if err == nil {
		t.Fatal("expected missing language configuration error")
	}
	want := "No configured languages in " + link
	if err.Error() != want {
		t.Fatalf("err = %q, want %q", err.Error(), want)
	}
}
