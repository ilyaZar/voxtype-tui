package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ilyaZar/voxtype-tui/internal/config"
)

func TestApplyWritesCycleAndResetsCurrent(t *testing.T) {
	t.Setenv("VOXTYPE_SKIP_RESTART", "1")
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.toml")
	cycleFile := filepath.Join(dir, "language_cycle.toml")
	text := strings.Join([]string{
		"[whisper]",
		"model = \"base\"",
		"language = \"de\"",
		"",
		"[[voxtype_tui.language]]",
		"code = \"en\"",
		"name = \"English\"",
		"label = \"EN\"",
		"model = \"base.en\"",
		"language = \"en\"",
		"",
		"[[voxtype_tui.language]]",
		"code = \"de\"",
		"name = \"German\"",
		"label = \"DE\"",
		"model = \"base\"",
		"language = \"de\"",
		"",
	}, "\n")
	if err := os.WriteFile(configFile, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Apply(config.Paths{ConfigFile: configFile, CycleFile: cycleFile}, []string{"en"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.ChangedCurrent || result.Current.Code != "en" {
		t.Fatalf("result = %#v", result)
	}

	cycle, err := os.ReadFile(cycleFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(cycle), `enabled = ["en"]`) {
		t.Fatalf("cycle not updated: %s", string(cycle))
	}

	updated, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(updated), `model = "base.en"`) || !strings.Contains(string(updated), `language = "en"`) {
		t.Fatalf("config not reset to en: %s", string(updated))
	}
}
