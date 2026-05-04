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

const (
	ansiRed   = "\x1b[31m"
	ansiReset = "\x1b[0m"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%sError: %v%s\n", ansiRed, err, ansiReset)
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
		if len(args) > 1 {
			return fmt.Errorf("unexpected version argument: %s", args[1])
		}
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
	cfg, _, err := loadConfig("toggle", args)
	if err != nil {
		return err
	}

	result, err := app.Toggle(cfg.Paths())
	if err != nil {
		return err
	}
	fmt.Println(app.ToggleMessage(result))
	return nil
}

func runSelected(args []string) error {
	cfg, _, err := loadConfig("selected", args)
	if err != nil {
		return err
	}

	paths := cfg.Paths()
	languages, err := config.ReadLanguages(paths.ConfigFile)
	if err != nil {
		return err
	}
	codes, err := config.ReadCycle(paths.CycleFile, languages)
	if err != nil {
		return err
	}
	fmt.Println(strings.Join(codes, "\n"))
	return nil
}

func runChoose(args []string) error {
	cfg, _, err := loadConfig("choose", args)
	if err != nil {
		return err
	}
	paths := cfg.Paths()

	languages, err := config.ReadLanguages(paths.ConfigFile)
	if err != nil {
		return err
	}
	selected, err := config.ReadCycle(paths.CycleFile, languages)
	if err != nil {
		return err
	}
	model := tui.New(selected, theme.Load(paths.ThemeFile), languages)
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
	fmt.Println(app.ApplyMessage(result))
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
	var (
		action string
		cfg    config.AppConfig
		err    error
	)
	if isRecordAction(args[0]) {
		action = args[0]
		cfg, _, err = loadConfig("record", args[1:])
	} else {
		var appConfig string
		var fs *flag.FlagSet
		appConfig, fs, err = parseConfig("record", args)
		if err != nil {
			return err
		}
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: voxtype-tui record [toggle|start|stop]")
		}
		action = fs.Arg(0)
		cfg, err = config.LoadAppConfig(appConfig)
	}
	if err != nil {
		return err
	}
	return app.Record(cfg, action)
}

func runPopup(args []string) error {
	cfg, configPath, err := loadConfig("popup", args)
	if err != nil {
		return err
	}
	return hypr.Popup(cfg, configPath)
}

func loadConfig(name string, args []string) (config.AppConfig, string, error) {
	appConfig, fs, err := parseConfig(name, args)
	if err != nil {
		return config.AppConfig{}, "", err
	}
	if err := rejectArgs(fs); err != nil {
		return config.AppConfig{}, "", err
	}
	cfg, err := config.LoadAppConfig(appConfig)
	if err != nil {
		return config.AppConfig{}, "", err
	}
	return cfg, effectiveConfigFile(appConfig), nil
}

func parseConfig(name string, args []string) (string, *flag.FlagSet, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	appConfig := fs.String("config-file", "", "voxtype-tui config.toml path")
	if err := fs.Parse(args); err != nil {
		return "", nil, err
	}
	return *appConfig, fs, nil
}

func effectiveConfigFile(path string) string {
	if path != "" {
		return path
	}
	if path = os.Getenv("VOXTYPE_TUI_CONFIG"); path != "" {
		return path
	}
	return config.DefaultConfigFile()
}

func rejectArgs(fs *flag.FlagSet) error {
	if fs.NArg() == 0 {
		return nil
	}
	return fmt.Errorf("unexpected %s argument: %s", fs.Name(), fs.Arg(0))
}

func isRecordAction(action string) bool {
	switch action {
	case "toggle", "start", "stop":
		return true
	default:
		return false
	}
}
