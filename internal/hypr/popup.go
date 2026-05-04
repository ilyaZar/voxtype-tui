package hypr

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/ilyaZar/voxtype-tui/internal/config"
)

type Monitor struct {
	ID              int       `json:"id"`
	X               int       `json:"x"`
	Y               int       `json:"y"`
	Width           int       `json:"width"`
	Reserved        []int     `json:"reserved"`
	Focused         bool      `json:"focused"`
	ActiveWorkspace Workspace `json:"activeWorkspace"`
}

type Workspace struct {
	ID int `json:"id"`
}

type Client struct {
	Address      string    `json:"address"`
	At           []int     `json:"at"`
	Size         []int     `json:"size"`
	Class        string    `json:"class"`
	InitialClass string    `json:"initialClass"`
	Title        string    `json:"title"`
	InitialTitle string    `json:"initialTitle"`
	Workspace    Workspace `json:"workspace"`
}

func Popup(cfg config.AppConfig, configFile string) error {
	if err := requireCommands("hyprctl", cfg.Popup.Terminal); err != nil {
		return err
	}

	monitor, err := focusedMonitor()
	if err != nil {
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}

	reservedTop := 0
	if len(monitor.Reserved) > 1 {
		reservedTop = monitor.Reserved[1]
	}
	localX := monitor.Width - cfg.Popup.Width - cfg.Popup.MarginX
	if localX < 0 {
		localX = cfg.Popup.MarginX
	}
	localY := reservedTop + cfg.Popup.MarginY
	globalX := monitor.X + localX
	globalY := monitor.Y + localY

	_ = dispatch("closewindow", "class:"+cfg.Popup.Class)
	_ = dispatch("closewindow", "title:"+cfg.Popup.Title)
	if err := waitForNoClient(cfg); err != nil {
		return err
	}

	rules := fmt.Sprintf(
		"[monitor %d; workspace %s silent; float; no_anim; size %d %d; move %d %d; tag -default-opacity; opacity %s]",
		monitor.ID,
		cfg.Popup.StageWorkspace,
		cfg.Popup.Width,
		cfg.Popup.Height,
		localX,
		localY,
		cfg.Popup.Opacity,
	)
	command := shellJoin([]string{
		cfg.Popup.Terminal,
		"--gtk-single-instance=false",
		"--class=" + cfg.Popup.Class,
		"--title=" + cfg.Popup.Title,
		"--working-directory=" + homeDir(),
		"--window-decoration=none",
		"--gtk-titlebar=false",
		"--window-padding-x=" + strconv.Itoa(cfg.Popup.TerminalPaddingX),
		"--window-padding-y=" + strconv.Itoa(cfg.Popup.TerminalPaddingY),
		"--window-width=" + strconv.Itoa(cfg.Popup.TerminalColumns),
		"--window-height=" + strconv.Itoa(cfg.Popup.TerminalRows),
		"-e",
		exe,
		"choose",
		"--config-file",
		configFile,
		"--window-x",
		strconv.Itoa(globalX),
		"--window-y",
		strconv.Itoa(globalY),
	})

	if err := dispatch("exec", rules+" "+command); err != nil {
		return err
	}

	client, err := waitForClient(cfg)
	if err != nil {
		return err
	}
	if err := placeStagedClient(cfg, client.Address, globalX, globalY); err != nil {
		return err
	}
	_ = dispatch("pin", "address:"+client.Address)
	if err := placeStagedClient(cfg, client.Address, globalX, globalY); err != nil {
		return err
	}
	_ = dispatch("focuswindow", "address:"+client.Address)
	return nil
}

func focusedMonitor() (Monitor, error) {
	data, err := hyprJSON("monitors")
	if err != nil {
		return Monitor{}, err
	}
	var monitors []Monitor
	if err := json.Unmarshal(data, &monitors); err != nil {
		return Monitor{}, fmt.Errorf("parse monitors: %w", err)
	}
	for _, monitor := range monitors {
		if monitor.Focused {
			return monitor, nil
		}
	}
	return Monitor{}, fmt.Errorf("focused Hypr monitor not found")
}

func waitForClient(cfg config.AppConfig) (Client, error) {
	for attempt := 0; attempt < cfg.Popup.WaitAttempts; attempt++ {
		client, ok, err := findClient(cfg)
		if err != nil {
			return Client{}, err
		}
		if ok {
			return client, nil
		}
		time.Sleep(time.Duration(cfg.Popup.WaitIntervalMS) * time.Millisecond)
	}
	return Client{}, fmt.Errorf("language selector window did not appear")
}

func waitForNoClient(cfg config.AppConfig) error {
	for attempt := 0; attempt < cfg.Popup.WaitAttempts; attempt++ {
		_, ok, err := findClient(cfg)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		time.Sleep(time.Duration(cfg.Popup.WaitIntervalMS) * time.Millisecond)
	}
	return fmt.Errorf("previous language selector window did not close")
}

func findClient(cfg config.AppConfig) (Client, bool, error) {
	data, err := hyprJSON("clients")
	if err != nil {
		return Client{}, false, err
	}
	var clients []Client
	if err := json.Unmarshal(data, &clients); err != nil {
		return Client{}, false, fmt.Errorf("parse clients: %w", err)
	}
	for _, client := range clients {
		if client.Class == cfg.Popup.Class || client.InitialClass == cfg.Popup.Class || client.Title == cfg.Popup.Title || client.InitialTitle == cfg.Popup.Title {
			return client, true, nil
		}
	}
	return Client{}, false, nil
}

func placeStagedClient(cfg config.AppConfig, address string, targetX int, targetY int) error {
	_ = dispatch("setprop", fmt.Sprintf("address:%s min_size %d %d", address, cfg.Popup.MinWidth, cfg.Popup.MinHeight))
	_ = dispatch("resizewindowpixel", fmt.Sprintf("exact %d %d,address:%s", cfg.Popup.Width, cfg.Popup.Height, address))

	placed := false
	for attempt := 0; attempt < cfg.Popup.PositionAttempts; attempt++ {
		time.Sleep(time.Duration(cfg.Popup.PositionSettleWaitMS) * time.Millisecond)
		client, ok, err := findClient(cfg)
		if err != nil {
			return err
		}
		if !ok || len(client.At) < 2 {
			continue
		}
		dx := targetX - client.At[0]
		dy := targetY - client.At[1]
		_ = dispatch("movewindowpixel", fmt.Sprintf("%d %d,address:%s", dx, dy, address))
		placed = true
	}
	if !placed {
		return fmt.Errorf("language selector geometry unavailable")
	}
	return nil
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
