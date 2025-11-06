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

### Go install (latest)
```sh
go install github.com/mabd-dev/reposcan@latest
```

Make sure $GOPATH/bin (or $HOME/go/bin) is in your $PATH.

### From source
```sh
git clone https://github.com/mabd-dev/reposcan.git
cd reposcan
go build -o reposcan ./cmd/reposcan
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
-o, --output string             # Output format: json|table|interactive|none (default "table")
-r, --root stringArray          # Root directory to scan (repeatable). Defaults to $HOME if unset in config. (default [$HOME])
  , --debug                     # Enable/Disable debug mode
```

Help
```sh
reposcan --help
```

More details on flags and config mapping can be found in [docs/cli-flags.md](docs/cli-flags.md).

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
- [x] Stdout Ouput in 3 formats: json, table, interactive, none
- [x] Read user customizable `config.toml` file
- [x] Export Report to json file
- [x] Support dirignore
- [x] Worker pool for speed
- [ ] Support git worktrees
- [ ] Perform git push/pull/fetch on repos
- [ ] Show branches with their states on each repo


## 🤝 Contributing
PRs, bug reports, and feature requests are welcome.
