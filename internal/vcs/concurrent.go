package vcs

import (
	"fmt"
	"sort"
	"sync"

	"github.com/mabd-dev/reposcan/pkg/report"
)

type repoResult struct {
	State    report.RepoState
	Warnings []string
	Include  bool
}

func GetRepoStatesConcurrent(
	repos []RepoPath,
	registry *Registry,
	maxWorkers int,
) ([]report.RepoState, []string) {
	if maxWorkers <= 0 {
		maxWorkers = 1
	}

	states := []report.RepoState{}
	warnings := []string{}

	jobs := make(chan RepoPath, maxWorkers*2)
	results := make(chan repoResult, maxWorkers*2)

	var wg sync.WaitGroup

	wg.Add(maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		go func() {
			defer wg.Done()

			for repo := range jobs {
				provider, ok := registry.Get(repo.Type)
				if !ok {
					results <- repoResult{
						Warnings: []string{
							fmt.Sprintf("No provider registered for vcs type=%q, path=%s", repo.Type, repo.Path),
						},
					}
					continue
				}

				state, repoWarnings := provider.CheckRepoState(repo.Path)
				results <- repoResult{
					State:    state,
					Warnings: repoWarnings,
					Include:  true,
				}
			}
		}()
	}

	go func() {
		for _, repo := range repos {
			jobs <- repo
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		if result.Include {
			states = append(states, result.State)
		}
		warnings = append(warnings, result.Warnings...)
	}

	sort.Slice(states, func(i, j int) bool { return states[i].Path < states[j].Path })

	return states, warnings
}
