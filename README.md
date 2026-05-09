# RepoScan

`reposcan` is a simple command-line tool written in Go that scans your filesystem for Git and jj repositories and reports their status.  
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

### Go install (latest)
```sh
go install github.com/mabd-dev/reposcan@latest
```

Make sure $GOPATH/bin (or $HOME/go/bin) is in your $PATH.

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
-f, --filter string             # Repository filter: all|dirty|uncommitted|unpushed|unpulled (default "dirty")
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

More details on flags and config mapping can be found in [docs/cli-flags.md](docs/cli-flags.md).

## VCS support

RepoScan currently discovers and reports on:

- Git repositories with a `.git` directory or worktree-style `.git` file.
- jj repositories with a `.jj` directory.

Reports include a `vcsType` field so JSON consumers and table users can distinguish Git and jj repositories. For jj repositories, RepoScan collects read-only state: repository name, current bookmark/change display, uncommitted file summaries, outgoing commits for tracked bookmarks, and incoming/unpulled counts based on already-fetched remote bookmark state.

Current jj limitations:

- TUI fetch, push, and pull keybindings are not active.
- jj fetch has a command wrapper but is not exposed through TUI actions yet.
- jj push and pull behavior is not enabled until per-operation semantics are defined.
- jj incoming/unpulled detection depends on tracked bookmarks and fetched remote bookmark state.
- jj remote status is simplified into a single synthetic status entry.
- JSON reports do not expose incoming commit details directly.
- TUI details show shared repository status fields, with limited jj-specific metadata.

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
- [x] Detect Git and jj repositories
- [x] Stdout Ouput in 3 formats: json, interactive, none
- [x] Read user customizable `config.toml` file
- [x] Export Report to json file
- [x] Support dirignore
- [x] Worker pool for speed
- [ ] Support git worktrees
- [ ] Perform git push/pull/fetch on repos
- [ ] Show branches with their states on each repo


## 🤝 Contributing
PRs, bug reports, and feature requests are welcome.
