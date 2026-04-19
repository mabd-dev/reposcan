# Changelog

All notable changes to this project will be documented in this file.

## [1.3.7] - 2025-11-06

### Added
- Show upstream status for each remote (git repos with multiple remotes)

### Changed
- Default output type changed from `table` to `interactive`
- `table` output type deprecated (will be removed in future releases)
- Made AGENTS.md source of truth, CLAUDE.md imports it (#17)

### Fixed
- Windows unit tests failing (#19)
- Build command documentation (#16)

---

## [1.3.6] - 2025-10-03

### Added
- `r` keybinding to rescan filesystem without restarting

### Changed
- Repository details now appear automatically when selecting a repo
- Filter input moved to footer when active
- Table expands to full terminal width
- Loop scrolling (navigate past last/first item wraps around)
- Introduced focus model stack pattern for managing interactive components
- Dedicated TUI overlay package
- Standalone repository details package
- Consistent TUI package structure (main.go, types.go, update.go, view.go)

### Fixed
- Filter cursor position lost when filtering repositories

---

## [1.3.5]

### Added
- Alerts system for errors, info, and warnings (auto-dismiss after 3 seconds)
- Added CLAUDE.md, removed AGENTS.md (chatgpt generated)
- Added workflows for Claude Code review action

---

## [1.3.4]

### Fixed
- Use go embed for colorscheme path

---

## [1.3.3]

### Fixed
- Removed colorscheme submodule, copied schemas to `internal/theme/base24-schemas/`

---

## [1.3.2]

### Fixed
- Colorschemes now properly embedded in binary (fixes `go install` issue)

---

## [1.3.1]

### Added
- File logger for debugging
- `version` command (`reposcan version`)
- User-configurable colorschemes (default: `catppuccin-mocha`)
- Keybindings overlay
- Footer moved to bottom of screen
- Repository table cursor indicator

---

## [1.3.0]

### Added
- Interactive terminal UI powered by Bubble Tea
- Clean table view of all repositories
- Real-time filtering and search
- Detailed repository information panel
- Responsive design that adapts to terminal size

### Changed
- **Breaking:** Changed `Output` and `JsonOutputPath` configuration format in `config.toml`:

```toml
# Old format
Output = "table"
JsonOutputPath = "/somewhere/nice"

# New format
[output]
type = "table"
jsonPath = "/somewhere/nice"
```
