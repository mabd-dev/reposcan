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
Generated at: 2025-08-31T14:12:45Z
Total repositories: 3  |  With uncommitted changes: 2

Repo                     Branch             Uncommitted  Path
---------------------------------------------------------------
empty                    main               0            /home/me/projects/empty
habitsss                 develop            2            /home/me/projects/habitsss
reposcan                 feature/cliFlags   1            /home/me/projects/reposcan

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
--only TYPE         # filter repos: all|uncommited
```

## ⚙️ Configuration
By default, `reposcan` looks for a config file in:
```sh
~/.config/reposcan/config.toml
```

Example
```toml
version = 1
roots = ["~/Code", "~/work"]

# Skip these directories (glob patterns)
dirIgnore = [
  "/node_modules/",
  "/.cache/",
  "/.local/"
]

only = "uncommited"

PrintStdOut = true
```
> You can still override everything via CLI flags.


## 🛣 Roadmap
- [x] Scan filesystem for repos
- [x] Detect uncommitted files
- [x] Detect uncommitted files
- [x] Read user customizable `config.toml` file
- [ ] Support unpushed + unpulled changes
- [ ] Support dirignore
- [ ] Export Report to json file
- [ ] Detect ahead/behind commits
- [ ] Worker pool for speed
- [ ] Support git worktrees
- [ ] Use cobra for better cli support


## 🤝 Contributing
PRs, bug reports, and feature requests are welcome.
