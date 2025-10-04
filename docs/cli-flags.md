# CLI Flags and Config Mapping

This document explains each CLI flag, its equivalent `config.toml` field, what it does, and examples with output snippets.

## Flags Overview

- `-r, --root PATH` (repeatable)
  - Config: `roots = ["/path1", "/path2"]`
  - Description: Directories to scan for Git repositories. Repeats to add multiple roots. Defaults to `$HOME` if unset.
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
  - Examples:
    - `reposcan --filter dirty`
    - `reposcan --filter uncommitted`
    - `reposcan --filter unpushed`
    - `reposcan --filter unpulled`

- `-o, --output TYPE`
  - Config: `Output = "table" | "json" | "none"`
  - Description: Select how results are printed to stdout.
    - `table`: human-friendly table + details.
    - `interactive`: human-friendly table + details with keymaps actions
    - `json`: machine-readable JSON object.
    - `none`: print nothing to stdout.
  - Example: `reposcan -o json`

- `--json-output-path DIR`
  - Config: `JsonOutputPath = "/path/to/reports"`
  - Description: Write a timestamped JSON report file to the directory. Folders are created if missing.
  - Example: `reposcan --json-output-path ./output-samples`

- `-w, --max-workers N`
  - Config: `maxWorkers = 16`
  - Description: Concurrency for git state checks when scanning many repos.
  - Example: `reposcan -w 16`

- `--debug true/false`
  - Config: `debug = true/false`
  - Description: Enable/disable logging mode. Log file will be in `~/.config/reposcan/logs/`
  - Example: `--debug=false` or `--debug` same as `--debug=true`

