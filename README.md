# RepoScan

`reposcan` is a simple command-line tool written in Go that scans your filesystem for Git repositories and reports their status.  
It helps you quickly find:

- Repositories with **uncommitted files**  
- Repositories with **unpushed commits** (ahead of upstream)  
- Repositories with **unpulled changes** (behind upstream)

It outputs results in both **human-friendly tables** and **machine-friendly JSON**, so you can use it interactively or integrate with scripts and future UIs.


🖼 Demo



https://github.com/user-attachments/assets/1c8370c6-3b94-4490-bc96-fc179ef14f1d




---

## ✨ Use cases

- **Daily cleanup**: See which projects have dirty working trees before switching tasks.
- **Context switch**: Know which repos still have unpushed commits before you leave for the day.
- **Housekeeping**: Find old repos you forgot to commit/push.
- **Automation**: Export JSON reports to integrate with dashboards or other tools.

---

## 📦 Installation

### Install script (recommended)

The easiest way to install `reposcan`. Detects your OS and architecture automatically and installs the latest release binary into a directory on your `$PATH`:

```sh
curl -fsSL https://raw.githubusercontent.com/mabd-dev/reposcan/main/install.sh | sh
```

Supports **linux/amd64**, **darwin/amd64**, and **darwin/arm64**.


#### Install environment variables

| env vars | Required | Default | Description |
|---|---|---|---|
| `VERSION` | false | latest | download a specific version |
| `ALIAS` | false | reposcan | specify binary name |

```sh
# with version
curl -fsSL https://raw.githubusercontent.com/mabd-dev/reposcan/main/install.sh | VERSION=v1.3.8 sh
```

```sh
# with alias
curl -fsSL https://raw.githubusercontent.com/mabd-dev/reposcan/main/install.sh | ALIAS=reposcan sh
```

```sh
# with both
curl -fsSL https://raw.githubusercontent.com/mabd-dev/reposcan/main/install.sh | VERSION=v1.3.8 ALIAS=reposcan sh
```


#### Migrating from `go install`

If you previously installed reposcan via `go install`, the binary lives in `$GOPATH/bin` (usually `~/go/bin/reposcan`). The curl installer puts the binary in a different location, so both can coexist silently — meaning the old one may take precedence in your `$PATH`.

To avoid this, remove the old binary first:

```sh
rm "$(which reposcan)"
```

Then install using the curl installer:

```sh
curl -fsSL https://raw.githubusercontent.com/mabd-dev/reposcan/main/install.sh | sh
```

### From source
```sh
git clone https://github.com/mabd-dev/reposcan.git
cd reposcan
go build -o reposcan .
```


## 🚀 Usage
Scan your home directory
```sh
reposcan -r $HOME
```

Multiple roots
```sh
reposcan -r ~/Code -r ~/work
```

Common flags
```sh
-d, --dirIgnore stringArray     # (default [$HOME])
-f, --filter string             # Repository filter: all|dirty|uncommitted|unpushed|unpulled|stash (default "dirty")
-h, --help                      # help for reposcan
    --json-output-path string   # Write scan report JSON files to this directory (optional)
-w, --max-workers int           # Number of concurrent git checks (default 8)
-o, --output string             # Output format: json|interactive|none (default "interactive")
-r, --root stringArray          # Root directory to scan (repeatable). Defaults to $HOME if unset in config. (default [$HOME])
  , --debug                     # Enable/Disable debug mode
```

Help
```sh
reposcan --help
```

More details on flags and config mapping can be found in [docs/cli-flags-and-configs.md](docs/cli-flags-and-configs.md).

## ⚙️ Configuration
By default, `reposcan` looks for a config file in: 
```sh
~/.config/reposcan/config.toml
```

Example
```toml
version = 1
debug = false

# directories to search for git repos inside
roots = ["~/Code", "~/work"]

only = "dirty"

# Count repos whose only local state is stashed work as dirty (default false).
# Only affects `only = "dirty"`; `only = "stash"` is unaffected.
countStashAsDirty = false

# Skip these directories (glob patterns)
dirIgnore = [
  "/node_modules/",
  "/.cache/",
  "/.local/"
]


[output]
type = "interactive"
jsonPath = "/somewhere/nice"


```
> You can still override everything via CLI flags.

check [sample/config.toml](sample/config.toml) for detailed configuration with examples


### Config lookup order
1. Load default values
1. Config in `~/.config/reposcan/config.toml` (if exists)
2. Cli flags (if exists)
Each step overrides the one before it


## 🛣 Roadmap
- [x] Scan filesystem for repos
- [x] Detect uncommitted files, unpushed commits and unpulled commits
- [x] Detect stashed changes
- [x] Stdout Ouput in 3 formats: json, interactive, none
- [x] Read user customizable `config.toml` file
- [x] Export Report to json file
- [x] Support dirignore
- [x] Worker pool for speed
- [x] Support git worktrees
- [ ] Perform git push/pull/fetch on repos
- [ ] Show branches with their states on each repo


## Telemetry

reposcan collects anonymous usage data to help understand how the tool is used and improve it over time.
You'll see a one-time notice about this on first run.

What is collected:
- `os` — operating system (linux, windows, darwin)
- `arch` — device cpu architecture 
- `tool-version` — tool version being used
- `ci` — whether the tool is running in a CI environment

and other tool specific cli-flags like `filter`, `output_format`, `repo_count`

Nothing personal is collected — no usernames, tokens, or file paths.
Events are sent to a [mixpanel](https://mixpanel.com/home/) (a third-party analytics service) and visible only to the maintainer.


### Disable telemetry

Add `--no-telemetry` when running the command. Or in `~/.config/reposcan/config.toml` add `no-telemetry = true` at the top of the file (check [sample.toml](sample/config.toml))


## 🤝 Contributing
PRs, bug reports, and feature requests are welcome.
