# Config Loading

How `reposcan` turns `config.toml` and CLI flags into the configuration a scan uses.

## Flow

```mermaid
flowchart TD
    A(["run reposcan, any command"]) --> B["init(): register flags only<br/>zero-value defaults"]
    B --> D{command}
    D -->|"version, update, help, completion"| E["run command<br/>(config not read)"]
    D -->|root| F["RunE: read config.toml<br/>(write the defaults first if the file is missing)"]
    F -->|read error| R1["return error"]
    F --> G["readFlags: apply only flags the user set"]
    G -->|"bad -o / -f value"| R1
    G --> V["validate the merged config"]
    V -->|invalid| R0["print issues"] --> R1
    V -->|valid| H{output}
    H -->|json| J["print report"]
    H -->|none| N["print nothing"]
    H -->|interactive| I["tui.Render"]
    I -->|error| R1
    I --> W
    J --> W
    N --> W["write JSON file if jsonPath is set"]
    W -->|error| R1
    R1 --> M["Execute(): print error, os.Exit(1)<br/>the only exit"]
```

- Only the root command reads `~/.config/reposcan/config.toml`.
  - `--help`, `version`, `update` and `completion` work even when the config is broken.
- Validation runs once, after flags are applied.
  - A flag can replace a bad config value, for example `reposcan -r ~/Code` when `roots` in the config points to a missing directory.
- Errors are returned to `Execute()`, which prints `Error: ...` and exits with status 1.

## Flag defaults and `--help`

Every flag that maps to a config key is registered with its zero value (`""`, `0`, `false`, empty list).

- `readFlags` applies a flag only when `cmd.Flags().Changed(name)` is true, so the registered value is never read.
- pflag omits `(default ...)` from `--help` for zero values, so help never shows a default that the user's config contradicts.
- See [cli-flags-and-configs.md](cli-flags-and-configs.md) for each flag's config key.

Flags that are not backed by config, such as `update --alias`, keep their real default.
