# RepoScan

`reposcan` is a simple command-line tool written in Go that scans your filesystem for Git repositories and reports their status.  
It helps you quickly find:

- Repositories with **uncommitted files**  
- Repositories with **unpushed commits** (ahead of upstream)  
- Repositories with **unpulled changes** (behind upstream)

It outputs results in both **human-friendly tables** and **machine-friendly JSON**, so you can use it interactively or integrate with scripts and future UIs.


🖼 Example output
```sh
Repo Scan Report
Generated at: 2025-08-31T08:44:54+03:00
Total repositories: 3  |  Dirty: 2

Repo                     Branch                State            Path
──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
empty                    main                  ⏳0  ↑0  ↓0      /home/me/projects/empty
habitsss                 master                ⏳2  ↑0  ↓2      /home/me/projects/habitsss
reposcan                 main                  ⏳1  ↑1  ↓0      /home/me/projects/reposcan


Details:

Repo: habitsss
Path: /home/me/projects/habitsss
  - internal/db/models.go
  - README.md

Repo: reposcan
Path: /home/me/projects/reposcan
  - api/handlers.go
```

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
go install github.com/you/reposcan/cmd/reposcan@latest
```

Make sure $GOPATH/bin (or $HOME/go/bin) is in your $PATH.

### From source
```sh
git clone https://github.com/MABD-dev/reposcan.git
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
-f, --filter string             # Repository filter: all|dirty (default "dirty")
-h, --help                      # help for reposcan
    --json-output-path string   # Write scan report JSON files to this directory (optional)
-w, --max-workers int           # Number of concurrent git checks (default 8)
-o, --output string             # Output format: json|table|none (default "table")
-r, --root stringArray          # Root directory to scan (repeatable). Defaults to $HOME if unset in config. (default [$HOME])
```

Help
```sh
reposcan --help
```

## ⚙️ Configuration
By default, `reposcan` looks for a config file in: 
```sh
~/.config/reposcan/config.toml
```

Example
```toml
version = 1

# directories to search for git repos inside
roots = ["~/Code", "~/work"]

# Skip these directories (glob patterns)
dirIgnore = [
  "/node_modules/",
  "/.cache/",
  "/.local/"
]

# options: 
#   1. `dirty`: git repos with un-commited changes or unpushed changes
#   2. `all`: all git repos
only = "dirty"

# print scan result to stdout. Options:
#   1. `json`: json object containing scan report struct
#   2. `table`: human readable representation of scan report
#   3. `none`: prints nothing
Output = "table"

# output scan reports to this folder. All nested folders will be created
# if they don't exist
JsonOutputPath = '/home/me/Documents/code/projects/Go/reposcan/output-samples'
```
> You can still override everything via CLI flags.

check [sample/config.toml](sample) for detailed configuration with examples

### Config lookup order
1. Load default values
1. Config in `~/.config/reposcan/config.toml` (if exists)
2. Cli flags (if exists)
Each step overrides the one before it


## 🛣 Roadmap
- [x] Scan filesystem for repos
- [x] Detect uncommitted files, unpushed commits and unpulled commits
- [x] Stdout Ouput in 3 formats: json, table, none
- [x] Read user customizable `config.toml` file
- [x] Export Report to json file
- [x] Support dirignore
- [ ] Worker pool for speed
- [ ] Support git worktrees
- [ ] Use cobra for better cli support


## 🤝 Contributing
PRs, bug reports, and feature requests are welcome.
