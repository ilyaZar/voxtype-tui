package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ilyaZar/voxtype-tui/internal/voxtype"
)

var quotedValueRE = regexp.MustCompile(`"([^"]+)"`)

func ReadCycle(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return voxtype.AllCodes(), nil
	}
	if err != nil {
		return nil, err
	}

	match := regexp.MustCompile(`(?ms)^\s*enabled\s*=\s*\[(.*?)\]`).FindSubmatch(data)
	if len(match) < 2 {
		return voxtype.AllCodes(), nil
	}

	codes := make([]string, 0)
	for _, value := range quotedValueRE.FindAllSubmatch(match[1], -1) {
		codes = append(codes, string(value[1]))
	}
	codes = voxtype.KnownCodes(codes)
	if len(codes) == 0 {
		return voxtype.AllCodes(), nil
	}
	return codes, nil
}

func WriteCycle(path string, codes []string) error {
	if len(codes) == 0 {
		return fmt.Errorf("select at least one language")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	quoted := make([]string, 0, len(codes))
	for _, code := range codes {
		quoted = append(quoted, fmt.Sprintf("%q", code))
	}
	text := "# Languages included in the Ctrl+F12 Voxtype cycle.\n"
	text += fmt.Sprintf("enabled = [%s]\n", strings.Join(quoted, ", "))
	return os.WriteFile(path, []byte(text), 0o644)
}
