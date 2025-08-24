package report

import (
	"time"
)

type GitRepo struct {
	Path string
}

type RepoState struct {
	Path            string   `json:"path"`
	Repo            string   `json:"repo"`
	Branch          string   `json:"branch"`
	UncommitedFiles []string `json:"uncommitedFiles"`
}

type ScanReport struct {
	Version     int         `json:"version"`
	RepoStates  []RepoState `json:"repoStates"`
	GeneratedAt time.Time   `json:"generatedAt"`
}
