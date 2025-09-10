# Repository Guidelines

## Project Structure & Module Organization
- `cmd/reposcan/`: CLI entrypoint (`main.go`, flags, filtering).
- `internal/`: Internal packages (config, git access, scan, render, utils, cli flag types).
- `pkg/report/`: Public types for reports consumed by renderers/outputs.
- `main.go`: Thin wrapper calling `cmd/reposcan.Run()`.
- `sample/`: Example config (`sample/config.toml`).
- `output-samples/`: Sample JSON outputs (keep sanitized for commits).

## Build, Test, and Development Commands
- Build CLI: `go build -o reposcan ./cmd/reposcan`
- Run locally: `go run ./cmd/reposcan --root $HOME --output table`
- Format: `go fmt ./...`  |  Vet: `go vet ./...`
- Tidy deps: `go mod tidy`
- Test (when present): `go test ./...`  |  Coverage: `go test -cover ./...`
- Release (CI): push a tag `vX.Y.Z` to trigger `.github/workflows/release.yml`.

## Coding Style & Naming Conventions
- Go 1.24+; use tabs (default `gofmt`) and idiomatic Go.
- Exported identifiers: `CamelCase`; packages: short, lower-case (`config`, `scan`).
- Filenames: lower_snake if helpful; keep packages cohesive.
- Run `go fmt` and `go vet` before pushing. Keep functions small and focused.

## Testing Guidelines
- Framework: standard `testing` package.
- Test files: `*_test.go` in the same package, e.g., `internal/scan/scan_test.go`.
- Names: `TestXxx(t *testing.T)`; table-driven tests preferred.
- Aim for coverage on config parsing, repo discovery, and git state helpers.
- Run: `go test ./...` (ensure tests are deterministic; avoid touching the network).

## Commit & Pull Request Guidelines
- Commits: imperative mood, concise scope (e.g., "scan: handle symlinks").
- Prefer small, focused PRs. Include context, reproduction steps, and before/after output when relevant.
- Link issues (`Fixes #123`) and include CLI examples or screenshots for UX changes.
- CI must pass; run format/vet/tests locally first.

## Security & Configuration Tips
- Do not commit secrets or personal paths in configs/output. Use `sample/config.toml` for examples.
- Default config path: `~/.config/reposcan/config.toml`. CLI flags override config values.
- Large scans: consider `--max-workers` to tune concurrency; be mindful of filesystem load.

