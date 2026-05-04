package hypr

import (
	"strings"
	"testing"

	"github.com/ilyaZar/voxtype-tui/internal/config"
)

func TestValidatePopupConfigRejectsNonGhosttyTerminal(t *testing.T) {
	cfg := config.DefaultAppConfig()
	cfg.Popup.Terminal = "alacritty"

	err := validatePopupConfig(cfg.Popup)
	if err == nil || !strings.Contains(err.Error(), "must be ghostty") {
		t.Fatalf("err = %v", err)
	}
}

func TestPopupCommandQuotesArguments(t *testing.T) {
	cfg := config.DefaultAppConfig().Popup
	cfg.Class = "class'withquote"
	command := popupCommand(cfg, "/tmp/voxtype-tui", "/tmp/config with space.toml")

	for _, want := range []string{
		"'--class=class'\\''withquote'",
		"'/tmp/voxtype-tui'",
		"'/tmp/config with space.toml'",
	} {
		if !strings.Contains(command, want) {
			t.Fatalf("command missing %q: %s", want, command)
		}
	}
}

func TestSameWorkspace(t *testing.T) {
	if !sameWorkspace(workspace{ID: 4, Name: "4"}, workspace{ID: 4, Name: "4"}) {
		t.Fatal("same named workspace did not match")
	}
	if sameWorkspace(workspace{ID: 3, Name: "3"}, workspace{ID: 4, Name: "4"}) {
		t.Fatal("different named workspaces matched")
	}
	if !sameWorkspace(workspace{ID: -98}, workspace{ID: -98}) {
		t.Fatal("same unnamed workspace did not match")
	}
	if !sameWorkspace(workspace{ID: 2}, workspace{ID: 2, Name: "2"}) {
		t.Fatal("same workspace with missing name did not match")
	}
}

func TestMatchesClient(t *testing.T) {
	cfg := config.DefaultAppConfig().Popup
	c := client{InitialClass: cfg.Class}
	if !matchesClient(c, cfg) {
		t.Fatalf("client did not match: %#v", c)
	}
}
