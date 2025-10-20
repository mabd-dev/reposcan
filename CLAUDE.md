# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

RepoScan is a Go CLI tool that scans filesystems for Git repositories and reports their status (uncommitted files, unpushed/unpulled commits). It provides multiple output formats: interactive TUI (using Bubble Tea), table, JSON, and can write reports to disk.

## Build & Test Commands

### Building
```bash
# Build the main binary
go build -o reposcan ./cmd/reposcan

# Install to $GOPATH/bin
go install github.com/mabd-dev/reposcan@latest
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/config
go test ./internal/scan

# Run tests with verbose output
go test -v ./...

# Run a specific test
go test -v -run TestName ./path/to/package
```

### Running
```bash
# Run with default settings (scans $HOME)
go run ./cmd/reposcan

# Run with custom root
go run ./cmd/reposcan -r ~/Code

# Run with specific output format
go run ./cmd/reposcan -o interactive
go run ./cmd/reposcan -o table
go run ./cmd/reposcan -o json
```

## Architecture

### Core Flow
1. **Configuration Loading** (`internal/config`): Loads defaults → reads `~/.config/reposcan/config.toml` → applies CLI flags (in that order of precedence)
2. **Repository Discovery** (`internal/scan`): Walks filesystem from root directories, applying `dirIgnore` glob patterns, identifying `.git` directories
3. **Git State Checking** (`internal/gitx`): Concurrently checks each repo using a worker pool (`gitxConcurrent.go`) to gather branch, uncommitted files, ahead/behind counts
4. **Filtering** (`cmd/reposcan/rootCmd.go`): Applies `OnlyFilter` (all/dirty/uncommitted/unpushed/unpulled) to determine which repos to include in output
5. **Rendering** (`internal/render`): Outputs results in chosen format (stdout table/json, interactive TUI, or file output)

### Key Packages

- **`cmd/reposcan`**: CLI entry point, Cobra command setup, flag parsing, orchestration of the scan→filter→render pipeline
- **`internal/config`**: Configuration types, validation, defaults, TOML loading. The `Config` struct in `types.go` is the central configuration object
- **`internal/scan`**: Filesystem walking with `filepath.WalkDir`, directory ignore matching using `doublestar` globs, git repo detection
- **`internal/gitx`**: Git operations via `exec.Command`. `gitFunctions.go` wraps individual git commands (status, branch, rev-list). `gitxConcurrent.go` implements worker pool pattern for parallel repo checking
- **`internal/render`**: Three render paths:
  - `stdout`: Plain table (using `charmbracelet/lipgloss`) or JSON output
  - `file`: Writes JSON reports to disk
  - `tui`: Interactive Bubble Tea interface with table, filtering, details panel, help popup, and git operations (push/pull/fetch)
- **`pkg/report`**: Public API types (`RepoState`, `ScanReport`) consumed by renderers and external tools
- **`internal/theme`**: Color schemes (base24 format) and lipgloss styles for TUI. `cmd/gen_schemes` generates theme definitions from YAML schemas

### Configuration Precedence
Values are merged in this order (later overrides earlier):
1. Defaults from `config.Defaults()`
2. TOML file at `~/.config/reposcan/config.toml`
3. CLI flags

### Concurrency Model
The `gitx.GetGitRepoStatesConcurrent` function uses a worker pool pattern:
- Creates buffered channels for jobs and results
- Spawns `maxWorkers` goroutines (default: 8)
- Each worker pulls repo paths from the jobs channel and checks git state
- Results are collected, sorted by path, and returned

### TUI Architecture (`internal/render/tui`)
Built with Bubble Tea (Elm architecture):
- **Model**: Contains `repostable.Table`, header, filter state, warnings, focused model tracker
- **Focused Model Pattern**: Different input modes (table navigation, filter text input, help popup) each implement `focusedModel` interface to handle updates and keybindings
- **Update Flow**: Messages route through focused model → update appropriate state → return new model + commands
- **View**: Composed vertically: header → body (table + optional filter/details) → footer (keybindings)
- **Git Operations**: TUI can trigger git push/pull/fetch via messages that execute git commands and update state

## Important Implementation Notes

### Git Detection
`scan.isGitRepo` checks for `.git` as either:
1. A directory (normal repos)
2. A file containing "gitdir:" (worktrees/submodules - partial support)

### Path Expansion
All root paths and dirIgnore patterns support environment variable expansion via `os.ExpandEnv`

### Filtering Logic
The `filter` function in `rootCmd.go` applies `OnlyFilter` after all repos are discovered and checked. This means all git operations happen regardless of filter—filtering only affects output.

### Error Handling in Scan
`scan.FindGitRepos` collects warnings (e.g., permission denied) but continues walking. Warnings are included in `ScanReport.Warnings`.

### Git Command Wrapper
`gitx.RunGitCommand` uses `git -C <dir>` to run commands in a specific directory without changing the process's working directory. Stderr is captured but only used for error detection—stdout is returned.

## Testing Patterns

Tests use standard Go testing:
- Config validation tests in `internal/config/*_test.go`
- Flag parsing tests in `cmd/reposcan/*_test.go`
- Scan behavior tests in `internal/scan/scan_test.go`
- File render tests in `internal/render/file/file_test.go`

When writing tests, prefer table-driven tests for multiple scenarios.
