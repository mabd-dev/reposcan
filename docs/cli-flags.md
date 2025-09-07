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
  - Config: `only = "dirty" | "all"`
  - Description: Filter which repositories to include in the report.
    - `dirty`: only repos with uncommitted changes or ahead/behind.
    - `all`: all repos discovered.
  - Example: `reposcan --filter dirty`

- `-o, --output TYPE`
  - Config: `Output = "table" | "json" | "none"`
  - Description: Select how results are printed to stdout.
    - `table`: human-friendly table + details.
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

## Example Runs

- Table output with two roots and ignore patterns
  ```sh
  reposcan -r ~/Code -r ~/work -d "**/node_modules/**" -o table
  ```
  Output (snippet):
  ```
  Repo Scan Report
  Generated at: 2025-08-31T22:00:00Z
  Total repositories: 3  |  Dirty: 2

  Repo        Branch   State        Path
  ----------------------------------------------
  example     main     ⏳1  ↑0  ↓0  /home/me/Code/example
  ```

- JSON output and write to file
  ```sh
  reposcan -r ~/Code -o json --json-output-path ./output-samples
  ```
  JSON (snippet):
  ```json
  {
    "Version": 1,
    "RepoStates": [ { "Repo": "example", "Branch": "main" } ]
  }
  ```

## Config File Example

Default lookup: `~/.config/reposcan/config.toml`

```toml
version = 1
roots = ["~/Code", "~/work"]
dirIgnore = ["**/node_modules/**", "/vendor/**"]
only = "dirty"
Output = "table"
JsonOutputPath = "~/reports/reposcan"
maxWorkers = 8
```

Notes:
- CLI flags override config values.
- Paths like `~`, `$HOME` are resolved using your home directory.
