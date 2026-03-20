package vcs

type Type string

const (
	TypeGit Type = "git"
	TypeJJ  Type = "jj"
)

type RepoPath struct {
	Path string
	Type Type
}
