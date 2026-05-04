package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ilyaZar/voxtype-tui/internal/app"
	"github.com/ilyaZar/voxtype-tui/internal/config"
	"github.com/ilyaZar/voxtype-tui/internal/hypr"
	"github.com/ilyaZar/voxtype-tui/internal/notify"
	"github.com/ilyaZar/voxtype-tui/internal/theme"
	"github.com/ilyaZar/voxtype-tui/internal/tui"
	"github.com/ilyaZar/voxtype-tui/internal/voxtype"
)

const usage = `usage: voxtype-tui <command> [options]

commands:
  choose    open the language-cycle selector
  language  manage language presets
  popup     open the placed language selector popup
  record    proxy voxtype record actions
  toggle    switch to the next selected language preset
  selected  print selected language codes
  version   print version
`

var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		notify.Send("critical", "5000", "Voxtype language preset", err.Error())
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		fmt.Print(usage)
		return nil
	}

	switch args[0] {
	case "choose":
		return runChoose(args[1:])
	case "language":
		return runLanguage(args[1:])
	case "popup":
		return runPopup(args[1:])
	case "record":
		return runRecord(args[1:])
	case "toggle":
		return runToggle(args[1:])
	case "selected":
		return runSelected(args[1:])
	case "version":
		fmt.Println(version)
		return nil
	case "help", "-h", "--help":
		fmt.Print(usage)
		return nil
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func runToggle(args []string) error {
	fs := flag.NewFlagSet("toggle", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	appConfig, _ := addConfigFlag(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, err := config.LoadAppConfig(*appConfig)
	if err != nil {
		return err
	}

	result, err := app.Toggle(cfg.Paths())
	if err != nil {
		return err
	}
	fmt.Printf("Switched to %s (cycle: %s; model: %s, lang: %s)\n",
		result.Current.Label,
		voxtype.Labels(result.Cycle),
		result.Current.Model,
		result.Current.Language,
	)
	return nil
}

func runSelected(args []string) error {
	fs := flag.NewFlagSet("selected", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	appConfig, _ := addConfigFlag(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, err := config.LoadAppConfig(*appConfig)
	if err != nil {
		return err
	}

	codes, err := config.ReadCycle(cfg.Paths().CycleFile)
	if err != nil {
		return err
	}
	fmt.Println(strings.Join(codes, "\n"))
	return nil
}

func runChoose(args []string) error {
	fs := flag.NewFlagSet("choose", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	appConfig, _ := addConfigFlag(fs)
	windowX := fs.Int("window-x", 0, "popup x coordinate supplied by wrapper")
	windowY := fs.Int("window-y", 0, "popup y coordinate supplied by wrapper")
	_ = windowX
	_ = windowY
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, err := config.LoadAppConfig(*appConfig)
	if err != nil {
		return err
	}
	paths := cfg.Paths()

	selected, err := config.ReadCycle(paths.CycleFile)
	if err != nil {
		return err
	}
	model := tui.New(selected, theme.Load(paths.ThemeFile))
	program := tea.NewProgram(model)
	final, err := program.Run()
	if err != nil {
		return err
	}
	finalModel, ok := final.(tui.Model)
	if !ok {
		return errors.New("unexpected TUI model")
	}
	if finalModel.Cancelled() || !finalModel.Done() {
		return nil
	}

	result, err := app.Apply(paths, finalModel.SelectedCodes())
	if err != nil {
		return err
	}
	detail := fmt.Sprintf(
		"Enabled: %s. Current stays %s (model: %s, lang: %s).",
		voxtype.Labels(result.Cycle), result.Current.Label, result.Current.Model, result.Current.Language,
	)
	if result.ChangedCurrent {
		detail = fmt.Sprintf(
			"Enabled: %s. Current reset to %s (model: %s, lang: %s).",
			voxtype.Labels(result.Cycle), result.Current.Label, result.Current.Model, result.Current.Language,
		)
	}
	fmt.Println(detail)
	return nil
}

func runLanguage(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: voxtype-tui language [toggle|selected|choose]")
	}
	switch args[0] {
	case "toggle":
		return runToggle(args[1:])
	case "selected":
		return runSelected(args[1:])
	case "choose":
		return runChoose(args[1:])
	default:
		return fmt.Errorf("unknown language command: %s", args[0])
	}
}

func runRecord(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: voxtype-tui record [toggle|start|stop]")
	}
	fs := flag.NewFlagSet("record", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	appConfig, _ := addConfigFlag(fs)
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	cfg, err := config.LoadAppConfig(*appConfig)
	if err != nil {
		return err
	}
	return app.Record(cfg, args[0])
}

func runPopup(args []string) error {
	fs := flag.NewFlagSet("popup", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	appConfig, defaultConfig := addConfigFlag(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, err := config.LoadAppConfig(*appConfig)
	if err != nil {
		return err
	}
	configPath := *appConfig
	if configPath == "" {
		configPath = os.Getenv("VOXTYPE_TUI_CONFIG")
	}
	if configPath == "" {
		configPath = defaultConfig
	}
	return hypr.Popup(cfg, configPath)
}

func addConfigFlag(fs *flag.FlagSet) (*string, string) {
	defaultConfig := config.DefaultConfigFile()
	path := fs.String("config-file", "", "voxtype-tui config.toml path")
	return path, defaultConfig
}
