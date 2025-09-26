package tui

func getRepoIndex(repoIds []string, id string) int {
	for i, x := range repoIds {
		if x == id {
			return i
		}
	}
	return -1
}

func deleteRepo(repoIds []string, index int) []string {
	return append(repoIds[:index], repoIds[index+1:]...)
}
