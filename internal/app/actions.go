package app

import (
	"fmt"
	"os"
	"os/exec"
	"slices"

	"github.com/ilyaZar/voxtype-tui/internal/config"
	"github.com/ilyaZar/voxtype-tui/internal/notify"
	"github.com/ilyaZar/voxtype-tui/internal/voxtype"
)

type Result struct {
	Current        voxtype.Language
	Cycle          []string
	Languages      []voxtype.Language
	ChangedCurrent bool
}

func Toggle(paths config.Paths) (Result, error) {
	languages, err := config.ReadLanguages(paths.ConfigFile)
	if err != nil {
		return Result{}, err
	}
	cycle, err := config.ReadCycle(paths.CycleFile, languages)
	if err != nil {
		return Result{}, err
	}
	current, ok, err := config.CurrentPreset(paths.ConfigFile, languages)
	if err != nil {
		return Result{}, err
	}

	targetCode := cycle[0]
	if ok {
		for index, code := range cycle {
			if code == current.Code {
				targetCode = cycle[(index+1)%len(cycle)]
				break
			}
		}
	}
	target, _ := voxtype.ByCode(targetCode, languages)
	if err := checkModels(paths.BaseModel, []voxtype.Language{target}); err != nil {
		return Result{}, err
	}
	if err := config.WritePreset(paths.ConfigFile, target); err != nil {
		return Result{}, err
	}
	if err := restart(); err != nil {
		return Result{}, err
	}

	result := Result{Current: target, Cycle: cycle, Languages: languages}
	notify.Send("normal", "5000", "Voxtype language preset", ToggleMessage(result))
	return result, nil
}

func Apply(paths config.Paths, selected []string) (Result, error) {
	languages, err := config.ReadLanguages(paths.ConfigFile)
	if err != nil {
		return Result{}, err
	}
	selected = voxtype.KnownCodes(selected, languages)
	if len(selected) == 0 {
		return Result{}, fmt.Errorf("select at least one language")
	}

	targets := languagesForCodes(languages, selected)
	if err := checkModels(paths.BaseModel, targets); err != nil {
		return Result{}, err
	}

	current, ok, err := config.CurrentPreset(paths.ConfigFile, languages)
	if err != nil {
		return Result{}, err
	}
	changedCurrent := false
	if !ok || !slices.Contains(selected, current.Code) {
		current = targets[0]
		changedCurrent = true
	}

	if err := config.WriteCycle(paths.CycleFile, selected); err != nil {
		return Result{}, err
	}
	if changedCurrent {
		if err := config.WritePreset(paths.ConfigFile, current); err != nil {
			return Result{}, err
		}
	}
	if err := restart(); err != nil {
		return Result{}, err
	}

	result := Result{Current: current, Cycle: selected, Languages: languages, ChangedCurrent: changedCurrent}
	notify.Send("normal", "5000", "Voxtype language cycle", ApplyMessage(result))
	return result, nil
}

func ToggleMessage(result Result) string {
	return fmt.Sprintf(
		"Switched to %s (cycle: %s; model: %s, lang: %s)",
		result.Current.Label,
		voxtype.Labels(result.Cycle, result.Languages),
		result.Current.Model,
		result.Current.Language,
	)
}

func ApplyMessage(result Result) string {
	status := "stays"
	if result.ChangedCurrent {
		status = "reset to"
	}
	return fmt.Sprintf(
		"Enabled: %s. Current %s %s (model: %s, lang: %s).",
		voxtype.Labels(result.Cycle, result.Languages),
		status,
		result.Current.Label,
		result.Current.Model,
		result.Current.Language,
	)
}

func checkModels(basePath string, targets []voxtype.Language) error {
	needsBase := false
	for _, target := range targets {
		if target.Model == "base" {
			needsBase = true
			break
		}
	}
	if !needsBase {
		return nil
	}
	if _, err := os.Stat(basePath); err != nil {
		return fmt.Errorf("non-English presets require base model, but model file is missing: %s", basePath)
	}
	return nil
}

func restart() error {
	if os.Getenv("VOXTYPE_SKIP_RESTART") == "1" {
		return nil
	}
	if _, err := exec.LookPath("systemctl"); err != nil {
		return fmt.Errorf("systemctl not found; restart voxtype manually")
	}
	if err := exec.Command("systemctl", "--user", "restart", "voxtype").Run(); err != nil {
		notify.Send("critical", "20000", "Voxtype language preset", "Voxtype restart failed; run: systemctl --user restart voxtype")
		return fmt.Errorf("voxtype restart failed: %w", err)
	}
	return nil
}

func languagesForCodes(known []voxtype.Language, codes []string) []voxtype.Language {
	languages := make([]voxtype.Language, 0, len(codes))
	for _, code := range codes {
		language, _ := voxtype.ByCode(code, known)
		languages = append(languages, language)
	}
	return languages
}

func Record(cfg config.AppConfig, action string) error {
	switch action {
	case "toggle", "start", "stop":
	default:
		return fmt.Errorf("unsupported record action: %s", action)
	}
	if cfg.Voxtype.Command == "" {
		return fmt.Errorf("voxtype command is not configured")
	}
	cmd := exec.Command(cfg.Voxtype.Command, "record", action)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("voxtype record %s failed: %w", action, err)
	}
	return nil
}
