Lazysf

Overview
- Lazysf is a Go TUI inspired by Lazygit that consolidates common Salesforce debugging workflows in a single, keyboard-driven interface. It shells out to the `sf` CLI for org data, Apex logs, and trace flags, and uses `ripgrep` for fast filtering.

Dependencies
- Go (1.22+)
- Salesforce CLI: `sf`
- ripgrep: `rg`
- Optional
  - `nvr` (neovim-remote) for opening logs in your existing Neovim instance
  - `jq` (useful for local debugging)

Build & Run
- Fetch deps and build:
  - `go mod tidy`
  - `go build -o lazysf ./cmd/lazysf`
- Optionally add to your PATH:
  - `mkdir -p ~/.config/bin && ln -s "$(pwd)/lazysf" ~/.config/bin/lazysf`
  - Ensure `~/.config/bin` is in PATH
- Run:
  - `lazysf` (or `go run ./cmd/lazysf` while developing)

Data & Paths
- Per-org data is stored under:
  - `~/.config/.lazysf/orgs/<alias>/logs/` (fallback to `~/.lazysf` if `~/.config` does not exist)
- Command logs and internal app logs are written to `./logs/` in your working directory.

Panels & Layout
- Panel 0: Status (top-left)
  - Shows alias → username, and Neovim server (when space allows)
- Panel 1: Traces (left column, top)
  - Manages trace flags for users (add, edit duration, delete)
  - Displays compact times: `DD HH:MM → DD HH:MM` and remaining time (`Left`)
- Panel 2: Logs / Filters (left column, middle; tabs)
  - Logs: navigate log list and preview on demand
  - Filters: manage include terms; applies `rg -n -C25` to current log
- Panel 3: Orgs (left column, bottom)
  - View/switch authenticated orgs; auth new orgs via web
- Panel 4: Main (right column)
  - Displays a loaded log; supports searching and scrolling
  - Note: loading a log (Space on Panel 2) does not change focus — use Tab to switch panels
- Panel 5: Command Log (bottom-right)
  - Shows recent commands with exit code, duration; right-aligned `?: help` hint

Expanded Modes
- Expand a panel to full width while keeping the helpbar visible:
  - Traces (1): Enter → expand/collapse
  - Logs (2): Enter → expand/collapse
  - Main (4): Enter → expand/collapse (keeps command log visible at bottom)
  - Cmd Log (5): Enter → expand/collapse
- Esc collapses any expanded panel.

Keybindings (summary)
- Global
  - `Tab`: cycle panels
  - `0..5`: focus panel
  - `?`: help modal
  - `Esc`: close modals / collapse expanded panel
  - `q` / `Ctrl-C`: quit

- Status (0)
  - Shows active org and NVIM server (short name)

- Traces (1)
  - `j`/`k`, `↑/↓`: move
  - `r`: refresh
  - `a`: add trace (prompt for user match; default 1h)
  - `e`: edit duration (accepts `30m`, `2h`, or minutes-only)
  - `d`: delete trace
  - `Enter`: expand/collapse
  - Note: If trace update fails due to exceeded debug log quota, lazysf prompts to delete all logs and retries.

- Logs / Filters (2)
- Logs tab
    - `j`/`k`, `↑/↓`, `PgUp/PgDn`, `Home/End`: navigate list
    - `Space`: load/preview selected log in Panel 4 (on demand)
      - Shows a spinner in Panel 4 while loading
      - Marks the opened log with `*` (e.g., `*>` when selected)
      - Does not auto-focus Panel 4; press `Tab` to switch
    - `Enter`: expand/collapse logs list
    - `D`: delete all Apex logs (confirmation)
  - Filters tab
    - `t`: toggle tabs
    - `a`: add filter
    - `e`: edit filter
    - `d`: delete filter
    - Filters apply `rg -n -C25` to current log in the main panel

- Orgs (3)
  - `j`/`k`, `↑/↓`, `PgUp/PgDn`, `Home/End`: navigate
  - `Space`: switch target org
  - `A`: auth new org via web

- Main (4)
  - `j`/`k`, `↑/↓`: scroll by one line
  - `PgUp/PgDn`, `Home/End`, `gg/G`: page/top/bottom
  - `U`/`D`: half-page jump up/down
  - `/`: search; `n`/`N`: next/prev
  - `o`: open in editor at current match (or top-of-view); sets `ft=log`
  - `S`: set Neovim server (updates status; used by `nvr` integration)
  - `Enter`: expand/collapse (keeps command log visible)
  - Hint: default message “Select a log with Space to load” appears until you load the first log

- Command Log (5)
  - `Enter`: expand/collapse
  - Shows recent commands with timing and status

Neovim Integration
- Opening logs in your existing Neovim instance uses `nvr`:
  - Preferred server resolution:
    1) `LAZYSF_NVIM_SERVER`
    2) `NVIM_LISTEN_ADDRESS`
    3) best-effort `nvr --remote-tab`
  - To target a specific Neovim instance:
    - Press `S` in Panel 4 and paste `:echo v:servername`
  - Log opens at current search hit line (or the top) with `ft=log` set.

Traces & Debug Log Quota
- If a trace create/update fails due to exceeding org debug log quota, lazysf prompts to delete all Apex logs (Tooling API) and retries automatically upon confirmation.

Filtering with ripgrep
- Filters tab builds an alternation pattern from your terms and runs:
  - `rg -n -C25 -e "term1|term2|..." <logfile>`
- Results are shown in Panel 4 with context blocks and search highlighting.

Troubleshooting
- Space on Logs tab not loading: ensure you’re on the Logs tab (press `t` to toggle when in Panel 2).
- `nvr` issues with multiple Neovim instances:
  - Set a specific server via `S` in Panel 4 or export `LAZYSF_NVIM_SERVER`.
- Command log truncation:
  - Press `Enter` on Panel 5 to expand; press `Esc` to collapse.

Development Notes
- Code Structure (high-level)
  - `cmd/lazysf/`: entrypoint
  - `internal/gui/`: views, layout, keybindings, modals
  - `internal/sf/`: thin wrappers around `sf` CLI (orgs, logs, traces, users)
  - `internal/rg/`: ripgrep helpers (future: parse `--json`)
  - `internal/utils/`: editor open, paths, logging
  - `internal/cmdlog/`: command log ring buffer + subscriptions
- Validate your changes by running:
  - `go build -o lazysf ./cmd/lazysf`
  - `go run ./cmd/lazysf`
