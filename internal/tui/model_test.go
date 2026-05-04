package tui

import (
	"reflect"
	"testing"

	"github.com/ilyaZar/voxtype-tui/internal/voxtype"
)

func TestFooterBlankLinesPinsFooterToBottom(t *testing.T) {
	if got := footerBlankLines(10, 4); got != 5 {
		t.Fatalf("blank lines = %d, want 5", got)
	}
}

func TestFooterBlankLinesKeepsFallbackGap(t *testing.T) {
	for _, height := range []int{0, 4, 5} {
		if got := footerBlankLines(height, 4); got != 1 {
			t.Fatalf("height %d blank lines = %d, want 1", height, got)
		}
	}
}

func TestLanguageMenusAlignShortcuts(t *testing.T) {
	languages := []voxtype.Language{
		{Code: "en", Name: "English", Label: "EN"},
		{Code: "de", Name: "German", Label: "DE"},
		{Code: "no", Name: "Norwegian", Label: "NO"},
		{Code: "ru", Name: "Russian", Label: "RU"},
	}
	menus := languageMenus(languages)
	want := []string{
		"English   (EN)",
		"German    (DE)",
		"Norwegian (NO)",
		"Russian   (RU)",
	}
	if !reflect.DeepEqual(menus, want) {
		t.Fatalf("menus = %#v, want %#v", menus, want)
	}
}
