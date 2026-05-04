# voxtype-tui

Terminal TUI for managing the Voxtype language cycle on Omarchy/Hyprland.

## Commands

```bash
voxtype-tui record toggle
voxtype-tui language toggle
voxtype-tui popup
voxtype-tui choose
voxtype-tui toggle
voxtype-tui selected
```

- `record toggle` proxies `voxtype record toggle`.
- `language toggle` switches to the next selected preset.
- `popup` opens or focuses the centered language selector terminal.
- `choose` opens a multi-select TUI and applies the selected language cycle.
- `toggle` is a compatibility alias for `language toggle`.
- `selected` prints enabled language codes, one per line.

## Build

Requires Go 1.23 or newer.

```bash
make test
make build
make install
```

`make install` writes `voxtype-tui` to `~/.local/bin` by default.

Install from GitHub:

```bash
GOBIN="$HOME/.local/bin" go install github.com/ilyaZar/voxtype-tui/cmd/voxtype-tui@latest
```

## Runtime Contract

- The binary owns Voxtype config edits, language-cycle state, model checks,
  record commands, popup launch/focus behavior, service restarts, and
  notifications.
- Runtime integration expects `voxtype`, `systemctl --user`, `hyprctl`, and
  Ghostty. `notify-send` is used when available.
- Hypr window rules own popup size and centering. If the selector is already on
  the current workspace, Shift+F12 focuses it; stale selectors on other
  workspaces are closed and relaunched.
- The default config path is `~/.config/voxtype-tui/config.toml`.
- The default integration paths match the repo-managed Omarchy overlay:
  - `~/.config/voxtype/config.toml`
  - `~/.config/voxtype/language_cycle.toml`
  - `~/.config/omarchy/current/theme/colors.toml`
- Language presets are discovered from `[[voxtype_tui.language]]` entries in
  the Voxtype config. If no presets are configured, commands fail with
  `No configured languages in ...`.

## Environment Overrides

| Variable                | Purpose                         |
|-------------------------|---------------------------------|
| `VOXTYPE_TUI_CONFIG`    | `voxtype-tui` config path       |
| `VOXTYPE_CONFIG`        | Voxtype `config.toml` path      |
| `VOXTYPE_CYCLE`         | language cycle TOML path        |
| `VOXTYPE_THEME_COLORS`  | Omarchy colors TOML path        |
| `VOXTYPE_BASE_MODEL`    | base model path for non-English |
| `VOXTYPE_SKIP_RESTART`  | skip `systemctl --user restart` |

## Keys

| Key                 | Action                |
|---------------------|-----------------------|
| `up`/`k`            | move up               |
| `down`/`j`          | move down             |
| `space`/`tab`/`x`   | toggle language       |
| `ctrl+a`            | select all languages  |
| `enter`             | apply                 |
| `esc`/`q`           | cancel without saving |
