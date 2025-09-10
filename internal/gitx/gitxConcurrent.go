package gitx

import (
	"github.com/MABD-dev/reposcan/pkg/report"
	"sort"
	"sync"
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
				// handle warnigns
				state, warnings := CheckRepoState(p)

				result := gitRepoResult{
					State:    state,
					Warnings: warnings,
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

	sort.Slice(states, func(i, j int) bool { return states[i].Path < states[j].Path })

	return states, warnings
}
