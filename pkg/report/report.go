package report

type RepoState struct {
	Path       string `json:"path"`
	Repo       string `json:"repo"`
	Branch     string `json:"branch"`
	Uncommited bool   `json:"uncommited"`
}

type ScanReport struct {
	Version     int         `json:"version"`
	RepoStates  []RepoState `json:"repoStates"`
	GeneratedAt time.Time   `json:"generatedAt"`
}
