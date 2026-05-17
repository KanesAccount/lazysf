# Lazysf

A Go terminal UI inspired by Lazygit for common Salesforce debugging workflows. Lazysf provides a keyboard-driven interface for orgs, Apex logs, trace flags, filtering, and editor integration while shelling out to the Salesforce CLI (`sf`) and `ripgrep` (`rg`).

## Requirements

- Go 1.22+
- Salesforce CLI (`sf`)
- ripgrep (`rg`)

Optional:

- `nvr` (`neovim-remote`) to open logs in an existing Neovim instance
- `jq` for local debugging

## Installation

```sh
go mod tidy
go build -o lazysf ./cmd/lazysf
```

Optionally add the binary to your `PATH`:

```sh
mkdir -p ~/.config/bin
ln -s "$(pwd)/lazysf" ~/.config/bin/lazysf
```

Ensure `~/.config/bin` is in your `PATH`.

## Usage

Run the built binary:

```sh
lazysf
```

Or run directly while developing:

```sh
go run ./cmd/lazysf
```

## Data and Logs

Per-org data is stored at:

```text
~/.config/.lazysf/orgs/<alias>/logs/
```

If `~/.config` does not exist, Lazysf falls back to `~/.lazysf`.

Command logs and internal application logs are written to `./logs/` in the current working directory.

## Layout

| Panel | Name | Description |
| --- | --- | --- |
| 0 | Status | Active org alias/username and Neovim server when available |
| 1 | Traces | Manage trace flags for users |
| 2 | Logs / Filters | Browse logs, load previews, and manage ripgrep filters |
| 3 | Orgs | View, switch, and authenticate Salesforce orgs |
| 4 | Main | Display loaded logs, search results, and filtered output |
| 5 | Command Log | Recent commands with exit code and duration |

## Keybindings

### Global

| Key | Action |
| --- | --- |
| `Tab` | Cycle panels |
| `0`-`5` | Focus panel |
| `?` | Open help |
| `Esc` | Close modal or collapse expanded panel |
| `q`, `Ctrl-C` | Quit |

### Traces (`1`)

| Key | Action |
| --- | --- |
| `j`/`k`, `↑`/`↓` | Move selection |
| `r` | Refresh |
| `a` | Add trace flag for a user, defaulting to 1 hour |
| `e` | Edit duration (`30m`, `2h`, or minutes-only) |
| `d` | Delete trace flag |
| `Enter` | Expand/collapse panel |

If a trace create/update fails because the debug log quota is exceeded, Lazysf prompts to delete all Apex logs and retries after confirmation.

### Logs / Filters (`2`)

Use `t` to switch between the Logs and Filters tabs.

#### Logs tab

| Key | Action |
| --- | --- |
| `j`/`k`, `↑`/`↓` | Move selection |
| `PgUp`/`PgDn`, `Home`/`End` | Navigate list |
| `Space` | Load selected log in the Main panel |
| `Enter` | Expand/collapse logs list |
| `D` | Delete all Apex logs after confirmation |

Loading a log does not move focus to the Main panel. Press `Tab` to switch panels. The opened log is marked with `*`.

#### Filters tab

| Key | Action |
| --- | --- |
| `a` | Add filter |
| `e` | Edit filter |
| `d` | Delete filter |

Filters are applied to the current log with:

```sh
rg -n -C25 -e "term1|term2|..." <logfile>
```

### Orgs (`3`)

| Key | Action |
| --- | --- |
| `j`/`k`, `↑`/`↓` | Move selection |
| `PgUp`/`PgDn`, `Home`/`End` | Navigate list |
| `Space` | Switch target org |
| `A` | Authenticate a new org via web |

### Main (`4`)

| Key | Action |
| --- | --- |
| `j`/`k`, `↑`/`↓` | Scroll one line |
| `PgUp`/`PgDn` | Page up/down |
| `Home`/`End`, `gg`/`G` | Jump to top/bottom |
| `U`/`D` | Half-page up/down |
| `/` | Search |
| `n`/`N` | Next/previous search result |
| `o` | Open in editor at current match or top of view |
| `S` | Set Neovim server |
| `Enter` | Expand/collapse panel, keeping Command Log visible |

Before a log is loaded, the Main panel displays: `Select a log with Space to load`.

### Command Log (`5`)

| Key | Action |
| --- | --- |
| `Enter` | Expand/collapse panel |

## Neovim Integration

Lazysf can open logs in an existing Neovim instance using `nvr`.

Server resolution order:

1. `LAZYSF_NVIM_SERVER`
2. `NVIM_LISTEN_ADDRESS`
3. Best-effort `nvr --remote-tab`

To target a specific Neovim instance, press `S` in the Main panel and paste the output of:

```vim
:echo v:servername
```

Logs open at the current search hit line, or the top of the current view, with `ft=log` set.

## Troubleshooting

- **Space does not load a log:** ensure Panel 2 is on the Logs tab. Press `t` to toggle tabs.
- **Neovim opens the wrong instance:** set a server with `S` in the Main panel or export `LAZYSF_NVIM_SERVER`.
- **Command output is truncated:** press `Enter` on Panel 5 to expand the Command Log.

## Development

Project structure:

```text
cmd/lazysf/       Entry point
internal/gui/     Views, layout, keybindings, and modals
internal/sf/      Wrappers around sf CLI commands
internal/rg/      ripgrep helpers
internal/utils/   Editor integration, paths, and logging
internal/cmdlog/  Command log ring buffer and subscriptions
```

Validate changes with:

```sh
go build -o lazysf ./cmd/lazysf
go run ./cmd/lazysf
```
