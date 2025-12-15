package gitx

import (
	"sort"
	"strings"
	"sync"

	"github.com/mabd-dev/reposcan/internal/logger"
	"github.com/mabd-dev/reposcan/pkg/report"
)

type gitRepoResult struct {
	State    report.RepoState
	Warnings []string
}

// GetGitRepoStatesConcurrent gathers RepoState for each path concurrently using
// up to maxWorkers goroutines. It returns states sorted by Path and any warnings.
func GetGitRepoStatesConcurrent(
	paths []string,
	maxWorkers int,
) ([]report.RepoState, []string) {

	if maxWorkers <= 0 {
		maxWorkers = 1
	}

	states := []report.RepoState{}
	warnings := []string{}

	jobs := make(chan string, maxWorkers*2)
	results := make(chan gitRepoResult, maxWorkers*2)

	var wg sync.WaitGroup

	wg.Add(maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		go func() {
			defer wg.Done()

			for p := range jobs {
				wtPaths, err := getWorktreesPaths(p)
				if err != nil {
					logger.Error("getWorktreesPaths() failed, ", logger.StringAttr("message", err.Error()))
					continue
				}

				worktrees := []report.Worktree{}
				repoWarnings := []string{}

				for _, wtPath := range wtPaths {
					worktree, warnings := GetWorktreeState(wtPath)

					if len(warnings) > 0 {
						repoWarnings = append(repoWarnings, warnings...)
					}
					worktrees = append(worktrees, worktree)
				}

				state, warnings := GetRepoState(p, worktrees)
				if len(warnings) > 0 {
					repoWarnings = append(repoWarnings, warnings...)
				}

				result := gitRepoResult{
					State:    state,
					Warnings: repoWarnings,
				}
				results <- result
			}
		}()
	}

	// Feed jobs in a separate goroutine, then close the jobs channel.
	go func() {
		for _, p := range paths {
			jobs <- p
		}
		close(jobs)
	}()

	// when all workers are done, close result
	go func() {
		wg.Wait()
		close(results)
	}()

	for x := range results {
		states = append(states, x.State)
		warnings = append(warnings, x.Warnings...)
	}

	sort.Slice(states, func(i, j int) bool { return strings.ToLower(states[i].Repo) < strings.ToLower(states[j].Repo) })

	return states, warnings
}
