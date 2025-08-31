# RepoScan

`reposcan` is a simple command-line tool written in Go that scans your filesystem for Git repositories and reports their status.  
It helps you quickly find:

- Repositories with **uncommitted files**  
- Repositories with **unpushed commits** (ahead of upstream)  
- Repositories with **unpulled changes** (behind upstream)  (todo)

It outputs results in both **human-friendly tables** and **machine-friendly JSON**, so you can use it interactively or integrate with scripts and future UIs.

🖼 Example output
```sh
Repo Scan Report
Generated at: 2025-08-31T08:44:54+03:00
Total repositories: 3  |  Dirty: 3

Repo                     Branch                    Not-Commited Ahead  Path
──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
empty                    main                      2            0      /home/me/projects/empty
habitsss                 master                    1            1     /home/me/projects/habitsss
reposcan                 main                      1            3      /home/me/projects/reposcan


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
reposcan --root $HOME
```

Multiple roots
```sh
reposcan --root ~/Code --root ~/work
```

Flags
```graphql
--root PATH         # add a directory to scan (repeatable)
--print-stdout      # print report to stdout in a human readable table
--only TYPE         # filter repos: all|dirty
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
#     1. `dirty`: git repos with un-commited changes or unpushed changes
#     2. `all`: all git repos
only = "dirty"

# if true, table with scan results will appear. Otherwise, only exist status will be returned 
PrintStdOut = true
```
> You can still override everything via CLI flags.

### Config lookup order
1.  check config in `~/.config/reposcan/config.toml`
2. check cli flags and override those in step 1


## 🛣 Roadmap
- [x] Scan filesystem for repos
- [x] Detect uncommitted files
- [x] Detect uncommitted files
- [x] Read user customizable `config.toml` file
- [x] Support ahead of remote
- [ ] Support behind remote
- [ ] Support dirignore
- [ ] Export Report to json file
- [ ] Detect ahead/behind commits
- [ ] Worker pool for speed
- [ ] Support git worktrees
- [ ] Use cobra for better cli support


## 🤝 Contributing
PRs, bug reports, and feature requests are welcome.
