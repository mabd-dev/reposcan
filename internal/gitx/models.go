package gitx

type GitRepo struct {
	Path     string
	RepoName string
	Branch   string
	Ahead    int
	Behind   int
}
