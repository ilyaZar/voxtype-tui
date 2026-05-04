package hypr

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ilyaZar/voxtype-tui/internal/config"
)

const (
	waitAttempts = 60
	waitInterval = 50 * time.Millisecond
)

type workspace struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type client struct {
	Address      string    `json:"address"`
	Class        string    `json:"class"`
	InitialClass string    `json:"initialClass"`
	Title        string    `json:"title"`
	InitialTitle string    `json:"initialTitle"`
	Workspace    workspace `json:"workspace"`
}

func Popup(cfg config.AppConfig, configFile string) error {
	popup := cfg.Popup
	if err := validatePopupConfig(popup); err != nil {
		return err
	}
	if err := requireCommands("hyprctl", popup.Terminal); err != nil {
		return err
	}

	workspace, err := activeWorkspace()
	if err != nil {
		return err
	}
	if client, ok, err := findClient(popup); err != nil {
		return err
	} else if ok {
		if sameWorkspace(client.Workspace, workspace) {
			return focusClient(client.Address)
		}
		if err := closeClient(client.Address); err != nil {
			return err
		}
		if err := waitForNoClient(popup); err != nil {
			return err
		}
	}

	if err := launchPopup(popup, configFile); err != nil {
		return err
	}
	client, err := waitForClient(popup)
	if err != nil {
		return err
	}
	return focusClient(client.Address)
}

func validatePopupConfig(cfg config.PopupConfig) error {
	if filepath.Base(cfg.Terminal) != "ghostty" {
		return fmt.Errorf("popup terminal must be ghostty; got %s", cfg.Terminal)
	}
	if cfg.Class == "" {
		return fmt.Errorf("popup class is not configured")
	}
	if cfg.Title == "" {
		return fmt.Errorf("popup title is not configured")
	}
	if cfg.TerminalColumns <= 0 || cfg.TerminalRows <= 0 {
		return fmt.Errorf("popup terminal size must be positive")
	}
	return nil
}

func launchPopup(cfg config.PopupConfig, configFile string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}
	return dispatch("exec", popupCommand(cfg, exe, configFile))
}

func popupCommand(cfg config.PopupConfig, exe string, configFile string) string {
	return shellJoin([]string{
		cfg.Terminal,
		"--gtk-single-instance=false",
		"--class=" + cfg.Class,
		"--title=" + cfg.Title,
		"--working-directory=" + homeDir(),
		"--window-decoration=none",
		"--gtk-titlebar=false",
		"--window-padding-x=" + strconv.Itoa(cfg.TerminalPaddingX),
		"--window-padding-y=" + strconv.Itoa(cfg.TerminalPaddingY),
		"--window-width=" + strconv.Itoa(cfg.TerminalColumns),
		"--window-height=" + strconv.Itoa(cfg.TerminalRows),
		"-e",
		exe,
		"choose",
		"--config-file",
		configFile,
	})
}

func activeWorkspace() (workspace, error) {
	data, err := hyprJSON("activeworkspace")
	if err != nil {
		return workspace{}, err
	}
	var current workspace
	if err := json.Unmarshal(data, &current); err != nil {
		return workspace{}, fmt.Errorf("parse active workspace: %w", err)
	}
	return current, nil
}

func waitForClient(cfg config.PopupConfig) (client, error) {
	return waitForClientState(cfg, true, "language selector window did not appear")
}

func waitForNoClient(cfg config.PopupConfig) error {
	_, err := waitForClientState(cfg, false, "previous language selector window did not close")
	return err
}

func waitForClientState(cfg config.PopupConfig, wantFound bool, timeout string) (client, error) {
	for attempt := 0; attempt < waitAttempts; attempt++ {
		found, ok, err := findClient(cfg)
		if err != nil {
			return client{}, err
		}
		if ok == wantFound {
			return found, nil
		}
		time.Sleep(waitInterval)
	}
	return client{}, errors.New(timeout)
}

func findClient(cfg config.PopupConfig) (client, bool, error) {
	data, err := hyprJSON("clients")
	if err != nil {
		return client{}, false, err
	}
	var clients []client
	if err := json.Unmarshal(data, &clients); err != nil {
		return client{}, false, fmt.Errorf("parse clients: %w", err)
	}
	for _, client := range clients {
		if matchesClient(client, cfg) {
			return client, true, nil
		}
	}
	return client{}, false, nil
}

func matchesClient(client client, cfg config.PopupConfig) bool {
	return client.Class == cfg.Class || client.InitialClass == cfg.Class || client.Title == cfg.Title || client.InitialTitle == cfg.Title
}

func sameWorkspace(a workspace, b workspace) bool {
	if a.Name != "" && b.Name != "" {
		return a.Name == b.Name
	}
	return a.ID == b.ID
}

func focusClient(address string) error {
	return dispatch("focuswindow", "address:"+address)
}

func closeClient(address string) error {
	return dispatch("closewindow", "address:"+address)
}

func hyprJSON(kind string) ([]byte, error) {
	cmd := exec.Command("hyprctl", kind, "-j")
	data, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("hyprctl %s -j failed: %w", kind, err)
	}
	return data, nil
}

func dispatch(name string, arg string) error {
	cmd := exec.Command("hyprctl", "dispatch", name, arg)
	if data, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("hyprctl dispatch %s failed: %w: %s", name, err, strings.TrimSpace(string(data)))
	}
	return nil
}

func requireCommands(names ...string) error {
	for _, name := range names {
		if _, err := exec.LookPath(name); err != nil {
			return fmt.Errorf("%s is required", name)
		}
	}
	return nil
}

func shellJoin(args []string) string {
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		quoted = append(quoted, shellQuote(arg))
	}
	return strings.Join(quoted, " ")
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return home
}
