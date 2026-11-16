# Changelog

All notable changes to this project will be documented in this file.


## Unreleased

- fix/config-startup-flow: `--help`, `version` and `update` work with a broken config.toml, `--help` no longer shows config values as defaults, and CLI flags can override invalid config values by @frittlechasm in [#104](https://github.com/mabd-dev/reposcan/pull/104)
- fix/partial-config-defaults: keys missing from `config.toml` fall back to their defaults instead of becoming empty, so a config with only `roots` works and the default `dirIgnore` patterns still apply by @mabd-dev in [#108](https://github.com/mabd-dev/reposcan/pull/108)
- docs/agents-behavior-change-rules: `AGENTS.md` rules to flag backward-incompatible changes before implementing, check shared code changes against `main`, and keep review fixes to the PR's scope by @mabd-dev in [#109](https://github.com/mabd-dev/reposcan/pull/109)
- fix/expand-tilde-paths: expand a leading `~` in `roots`, `dirIgnore` and `jsonPath`, so the README example config works and the default `~/Library/**` ignore matches. `--json-output-path` with `~/..` or `$VAR/..` now writes to the right folder ([#111](https://github.com/mabd-dev/reposcan/issues/111)) by @mabd-dev in [#107](https://github.com/mabd-dev/reposcan/pull/107)
- ci/smoke-test: CI runs `scripts/smoke-test.sh` on Linux and macOS in place of the `go test` steps, so the built binary is checked end to end by @mabd-dev in [#106](https://github.com/mabd-dev/reposcan/pull/106)
- fix/docs-links-and-release-go-version: update `release.yml` github action go version to use the one in go.mod by @mabd-dev in [#105](https://github.com/mabd-dev/reposcan/pull/105)
- fix/jj-branch-display-after-fetch: jj repos count ahead/behind like git's `HEAD` against every remote, keep showing the bookmark after a fetch moves it past `@` when no other bookmark sits below `@`, and show `x` when a remote has no matching bookmark by @frittlechasm in [#98](https://github.com/mabd-dev/reposcan/pull/98)
- feat/staleDays: add `--stale-days` / `staleDays` to show only repos with no local activity for N days by @mabd-dev in [#79](https://github.com/mabd-dev/reposcan/pull/79)
- feat/jjSupport: add Jujutsu (jj) repo support alongisde git repos by @frittlechasm in [#62](https://github.com/mabd-dev/reposcan/pull/62)
- fix/crashInRepoDetails: repo details package crashes if repo state is nil - happens if scan result has no repos by @mabd-dev in [#61](https://github.com/mabd-dev/reposcan/pull/61)
- feat/colorSchemeSwitcher: change app color scheme from bunch of predefined schemes by @mabd-dev in [#60](https://github.com/mabd-dev/reposcan/pull/60)
- feat/showStashes: change repodetails view to show repo changes and stashes in tabbed view by @mabd-dev in [#57](https://github.com/mabd-dev/reposcan/pull/57)
- feat/migrateBubbleTeaToV2: migrate bubble tea and its related packages to v2 by @mabd-dev in [#56](https://github.com/mabd-dev/reposcan/pull/56)
- feat/updateCmd: added new `reposcan update` command that updates binary to the latest version by @mabd-dev in [#55](https://github.com/mabd-dev/reposcan/pull/55)
- feat/jjSupport: add configurable VCS column by @frittlechasm in [#51](https://github.com/mabd-dev/reposcan/pull/51) 


## [1.4.1] - 2026-09-09

### Fix
- fix/repoDetailsRuntimeError: uncommited files view panic when `maxUncommitedFilesToShow` is negative  by @mabd-dev in [#75](https://github.com/mabd-dev/reposcan/pull/75)

---

## [1.4.0] - 2026-05-31

### Changes

- feat/handleEmptyRepoTable: Display empty message when repos table is empty @kings9527 in [#42](https://github.com/mabd-dev/reposcan/pull/42)
- feat/colorUncommitedFiles: Get color based on file git status by @mabd-dev in [#50](https://github.com/mabd-dev/reposcan/pull/50)
- feat/detect-stashes: Detect stashed changes as a repo state; adds `stash` filter and opt-in `countStashAsDirty` config by @nilp0inter in [#47](https://github.com/mabd-dev/reposcan/pull/47)
- feat/supportWorktree: Detect git worktrees and submodules (`.git` file with `gitdir:` pointer) during repo scan by @mabd-dev in [#43](https://github.com/mabd-dev/reposcan/pull/43)

---

## [1.3.8] - 2026-05-21

### Added

- feat(telemetry): send anonymous telemetry usage by @mabd-dev in [#39](https://github.com/mabd-dev/reposcan/pull/39)
- feat(release): release workflow and install.sh by @AnthonyGeorgeAZ in [#30](https://github.com/mabd-dev/reposcan/pull/30)
- feat(analytics): add Analytics interface with Mixpanel and stdout backends by @mvanhorn in [#27](https://github.com/mabd-dev/reposcan/pull/27)
- feat(changelog): add CHANGELOG.md by @vgauraha62 in [#26](https://github.com/mabd-dev/reposcan/pull/26)

### Fixed

- fix: hide redundant remote name in state column by @mvanhorn in [#23](https://github.com/mabd-dev/reposcan/pull/23)

---

## [1.3.8-alpha] - 2026-04-27

### Added

- Feat/release workflow and install.sh by @AnthonyGeorgeAZ in [#30](https://github.com/mabd-dev/reposcan/pull/30)
- CHANGELOG.md added by @vgauraha62 in [#26](https://github.com/mabd-dev/reposcan/pull/26)
- feat(analytics): add Analytics interface with Mixpanel and stdout backends by @mvanhorn in [#27](https://github.com/mabd-dev/reposcan/pull/27)

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

