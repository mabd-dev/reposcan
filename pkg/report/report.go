package report

import (
	"time"
)

type RepoState struct {
	ID              string   `json:"id"`
	Path            string   `json:"path"`
	Repo            string   `json:"repo"`
	Branch          string   `json:"branch"`
	UncommitedFiles []string `json:"uncommitedFiles"`
	Ahead           int      `json:"ahead"`
	Behind          int      `json:"behind"`
}

type ScanReport struct {
	Version     int         `json:"version"`
	RepoStates  []RepoState `json:"repoStates"`
	GeneratedAt time.Time   `json:"generatedAt"`
}
