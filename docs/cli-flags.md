# CLI Flags and Config Mapping

This document explains each CLI flag, its equivalent `config.toml` field, what it does, and examples with output snippets.

## Flags Overview

- `-r, --root PATH` (repeatable)
  - Config: `roots = ["/path1", "/path2"]`
  - Description: Directories to scan for Git and jj repositories. Repeats to add multiple roots. Defaults to `$HOME` if unset.
  - Example:
    - CLI: `reposcan -r ~/Code -r ~/work`
    - TOML:
      ```toml
      roots = ["~/Code", "~/work"]
      ```

- `-d, --dirIgnore GLOB` (repeatable)
  - Config: `dirIgnore = ["**/node_modules/**", "/vendor/**"]`
  - Description: Glob patterns to skip while walking directories. Supports doublestar patterns.
  - Example:
    - CLI: `reposcan -d "**/node_modules/**" -d "/vendor/**"`
    - TOML:
      ```toml
      dirIgnore = ["**/node_modules/**", "/vendor/**"]
      ```

- `-f, --filter TYPE`
  - Config: `only = "dirty" | "all" | "uncommitted" | "unpushed" | "unpulled"`
  - Description: Filter which repositories to include in the report.
    - `dirty`: any of uncommitted files, ahead, or behind.
    - `uncommitted`: only repos with uncommitted files.
    - `unpushed`: only repos with commits ahead of upstream.
    - `unpulled`: only repos with commits behind upstream.
    - `all`: all repos discovered.
  - jj note: `unpushed` uses outgoing commits for tracked bookmarks. `unpulled` uses incoming commits inferred from already-fetched remote bookmark state.
  - Examples:
    - `reposcan --filter dirty`
    - `reposcan --filter uncommitted`
    - `reposcan --filter unpushed`
    - `reposcan --filter unpulled`

- `-o, --output TYPE`
  - Config: `output.type = "json" | "interactive" | "none"`
  - Description: Select how results are printed to stdout.
    - `interactive`: human-friendly table + details with keymaps actions
    - `json`: machine-readable JSON object.
    - `none`: print nothing to stdout.
  - Example: `reposcan -o json`
  - JSON note: repository entries include `vcsType`. jj entries currently expose outgoing commits in `remoteStatus[].outgoingCommits`; incoming commit details are used for the behind count but are not exposed directly.

- `--json-output-path DIR`
  - Config: `output.jsonPath = "/path/to/reports"`
  - Description: Write a timestamped JSON report file to the directory. Folders are created if missing.
  - Example: `reposcan --json-output-path ./output-samples`

- `-w, --max-workers N`
  - Config: `maxWorkers = 16`
  - Description: Concurrency for VCS state checks when scanning many repos.
  - Example: `reposcan -w 16`

- `--debug true/false`
  - Config: `debug = true/false`
  - Description: Enable/disable logging mode. Log file will be in `~/.config/reposcan/logs/`
  - Example: `--debug=false` or `--debug` same as `--debug=true`

## jj support notes

RepoScan supports read-only jj repository reporting. It discovers repositories with `.jj` directories and reports shared fields such as repo name, path, current bookmark/change display, uncommitted files, and remote status.

Current jj limitations:

- TUI fetch, push, and pull keybindings are inactive.
- `jj git fetch` has an internal command wrapper, but fetch is not exposed through TUI actions yet.
- jj push and pull are not enabled until their bookmark and pull semantics are defined.
- jj unpulled detection depends on tracked bookmarks and already-fetched remote bookmark state.
- jj remote status is currently represented as one synthetic status entry rather than remote/bookmark-level entries.
