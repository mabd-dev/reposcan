# Changelog

All notable changes to this project will be documented in this file.


## Unreleased


### 🐛 Bug Fixes

- fix: hide redundant remote name in state column by @mvanhorn in [#23](https://github.com/mabd-dev/reposcan/pull/23)

---

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

## [1.3.5] - 2025-10-20

### Added
- Alerts system for errors, info, and warnings (auto-dismiss after 3 seconds)
- Added CLAUDE.md, removed AGENTS.md (chatgpt generated)
- Added workflows for Claude Code review action

---

## [1.3.4] - 2025-10-11

### Fixed
- Use go embed for colorscheme path

---

## [1.3.3] - 2025-10-11

### Fixed
- Removed colorscheme submodule, copied schemas to `internal/theme/base24-schemas/`

---

## [1.3.2] - 2025-10-11

### Fixed
- Colorschemes now properly embedded in binary (fixes `go install` issue)

---

## [1.3.1] - 2025-10-11

### Added
- File logger for debugging
- `version` command (`reposcan version`)
- User-configurable colorschemes (default: `catppuccin-mocha`)
- Keybindings overlay
- Footer moved to bottom of screen
- Repository table cursor indicator

---

## [1.3.0] - 2025-10-03

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

---


## [1.2.0] - 2025-09-10

### 🚀 Features
- **Scan folders faster** ([#8](#8))
- **Cobra CLI integration** ([#9](#9))
- **Unit tests** ([#10](#10))

### ✨ Improvements
- Add `ahead+behind` to the same output column (`16254a49be`)
- Group `uncommitted`, `unpushed`, and `unpulled` into a single output column (`da3eb4f2ff`)
- Save visited directories when scanning roots (`939b9ee37a`)
- Added `uncommitted`, `unpushed`, and `unpulled` filter options (`d4407473a4`)
- Added more `dirignore` entries to sample `config.yaml` (`d0fa8b9890`)

### 📝 Documentation
- Added docs for each CLI flag (`99fc0e7567`)
- Added docs to functions (`69b783ac11`)
- Added `agents.md` (`c942326954`)
- Updated CLI output example (`12f769ca02`)
- Updated README (`1531ef4d9e`, `4a0ff630c0`)

### 🐛 Fixes
- Fix arrow position (`31db6585c5`)
- Fix typos in README (`12f769ca02`)

