package vcs

type Type string

const (
	TypeGit Type = "git"
	TypeJJ  Type = "jj"
)

type RepoInfo struct {
	Path string
	Type Type
}
