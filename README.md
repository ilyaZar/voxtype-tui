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
- `popup` opens the placed language selector terminal.
- `choose` opens a multi-select TUI and applies the selected language cycle.
- `toggle` is a compatibility alias for `language toggle`.
- `selected` prints enabled language codes, one per line.

## Build

```bash
make tidy
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
  record commands, popup placement, service restarts, and notifications.
- `popup` owns monitor geometry and Ghostty placement. It stages the terminal on
  a hidden special workspace before reveal to avoid visible warping.
- The default config path is `~/.config/voxtype-tui/config.toml`.
- The default integration paths match the repo-managed Omarchy overlay:
  - `~/.config/voxtype/config.toml`
  - `~/.config/voxtype/language_cycle.toml`
  - `~/.config/omarchy/current/theme/colors.toml`

## Environment Overrides

| Variable                | Purpose                         |
|-------------------------|---------------------------------|
| `VOXTYPE_TUI_CONFIG`    | `voxtype-tui` config path       |
| `VOXTYPE_CONFIG`        | Voxtype `config.toml` path      |
| `VOXTYPE_CYCLE`         | language cycle TOML path        |
| `VOXTYPE_THEME_COLORS`  | Omarchy colors TOML path        |
| `VOXTYPE_BASE_MODEL`    | base model path for DE/RU       |
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
