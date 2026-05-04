package app

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/ilyaZar/voxtype-tui/internal/config"
	"github.com/ilyaZar/voxtype-tui/internal/notify"
	"github.com/ilyaZar/voxtype-tui/internal/voxtype"
)

type Result struct {
	Current        voxtype.Language
	Cycle          []string
	ChangedCurrent bool
}

func Toggle(paths config.Paths) (Result, error) {
	cycle, err := config.ReadCycle(paths.CycleFile)
	if err != nil {
		return Result{}, err
	}
	current, ok, err := config.CurrentPreset(paths.ConfigFile)
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
	target, _ := voxtype.ByCode(targetCode)
	if err := CheckModels(paths.BaseModel, []voxtype.Language{target}); err != nil {
		return Result{}, err
	}
	if err := config.WritePreset(paths.ConfigFile, target); err != nil {
		return Result{}, err
	}
	if err := Restart(); err != nil {
		return Result{}, err
	}

	message := fmt.Sprintf(
		"Switched to %s (cycle: %s; model: %s, lang: %s)",
		target.Label,
		voxtype.Labels(cycle),
		target.Model,
		target.Language,
	)
	notify.Send("normal", "5000", "Voxtype language preset", message)
	return Result{Current: target, Cycle: cycle}, nil
}

func Apply(paths config.Paths, selected []string) (Result, error) {
	selected = voxtype.KnownCodes(selected)
	if len(selected) == 0 {
		return Result{}, fmt.Errorf("select at least one language")
	}

	targets := make([]voxtype.Language, 0, len(selected))
	for _, code := range selected {
		language, _ := voxtype.ByCode(code)
		targets = append(targets, language)
	}
	if err := CheckModels(paths.BaseModel, targets); err != nil {
		return Result{}, err
	}

	current, ok, err := config.CurrentPreset(paths.ConfigFile)
	if err != nil {
		return Result{}, err
	}
	changedCurrent := false
	if !ok || !contains(selected, current.Code) {
		current = targets[0]
		if err := config.WritePreset(paths.ConfigFile, current); err != nil {
			return Result{}, err
		}
		changedCurrent = true
	}

	if err := config.WriteCycle(paths.CycleFile, selected); err != nil {
		return Result{}, err
	}
	if err := Restart(); err != nil {
		return Result{}, err
	}

	detail := fmt.Sprintf(
		"Enabled: %s. Current stays %s (model: %s, lang: %s).",
		voxtype.Labels(selected), current.Label, current.Model, current.Language,
	)
	if changedCurrent {
		detail = fmt.Sprintf(
			"Enabled: %s. Current reset to %s (model: %s, lang: %s).",
			voxtype.Labels(selected), current.Label, current.Model, current.Language,
		)
	}
	notify.Send("normal", "5000", "Voxtype language cycle", detail)
	return Result{Current: current, Cycle: selected, ChangedCurrent: changedCurrent}, nil
}

func CheckModels(basePath string, targets []voxtype.Language) error {
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
		return fmt.Errorf("DE/RU requires base model, but model file is missing: %s", basePath)
	}
	return nil
}

func Restart() error {
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

func contains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
