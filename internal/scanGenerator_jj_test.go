package internal

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/mabd-dev/reposcan/internal/vcs"
)

func requireJJAndGit(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj binary not available")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}
}

func initScanJJRepo(t *testing.T, root string, name string) string {
	t.Helper()

	repoPath := filepath.Join(root, name)
	if err := exec.Command("jj", "git", "init", repoPath).Run(); err != nil {
		t.Fatalf("jj git init: %v", err)
	}

	return repoPath
}

type scanTrackedJJRepo struct {
	SeedPath string
	WorkPath string
}

func initScanTrackedJJRepo(t *testing.T) scanTrackedJJRepo {
	t.Helper()

	root := t.TempDir()
	remotePath := filepath.Join(root, "remote.git")
	seedPath := filepath.Join(root, "seed")
	workPath := filepath.Join(root, "work")

	if err := exec.Command("git", "init", "--bare", remotePath).Run(); err != nil {
		t.Fatalf("git init --bare: %v", err)
	}
	if err := exec.Command("git", "clone", remotePath, seedPath).Run(); err != nil {
		t.Fatalf("git clone: %v", err)
	}
	if err := exec.Command("git", "-C", seedPath, "config", "user.name", "test").Run(); err != nil {
		t.Fatalf("git config user.name: %v", err)
	}
	if err := exec.Command("git", "-C", seedPath, "config", "user.email", "test@example.com").Run(); err != nil {
		t.Fatalf("git config user.email: %v", err)
	}
	if err := os.WriteFile(filepath.Join(seedPath, "README.md"), []byte("one\n"), 0o644); err != nil {
		t.Fatalf("write seed README: %v", err)
	}
	if err := exec.Command("git", "-C", seedPath, "add", "README.md").Run(); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if err := exec.Command("git", "-C", seedPath, "commit", "-m", "initial").Run(); err != nil {
		t.Fatalf("git commit: %v", err)
	}
	if err := exec.Command("git", "-C", seedPath, "branch", "-M", "main").Run(); err != nil {
		t.Fatalf("git branch -M main: %v", err)
	}
	if err := exec.Command("git", "-C", seedPath, "push", "origin", "main").Run(); err != nil {
		t.Fatalf("git push origin main: %v", err)
	}
	if err := exec.Command("jj", "git", "clone", remotePath, workPath).Run(); err != nil {
		t.Fatalf("jj git clone: %v", err)
	}

	return scanTrackedJJRepo{
		SeedPath: seedPath,
		WorkPath: workPath,
	}
}

func jjScanConfig(root string, only config.OnlyFilter) config.Config {
	cfg := config.Defaults()
	cfg.Roots = []string{root}
	cfg.DirIgnore = nil
	cfg.Only = only
	cfg.MaxWorkers = 1
	return cfg
}

func scanOneJJRepo(t *testing.T, root string, only config.OnlyFilter) {
	t.Helper()

	scanReport := GenerateScanReport(jjScanConfig(root, only))
	if len(scanReport.Warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", scanReport.Warnings)
	}
	if len(scanReport.RepoStates) != 1 {
		t.Fatalf("expected one repo for only=%s, got %d: %#v", only, len(scanReport.RepoStates), scanReport.RepoStates)
	}
	if scanReport.RepoStates[0].VCSType != string(vcs.TypeJJ) {
		t.Fatalf("expected jj repo, got vcsType=%q", scanReport.RepoStates[0].VCSType)
	}
}

func TestGenerateScanReportFiltersJJUncommittedRepos(t *testing.T) {
	requireJJAndGit(t)

	root := t.TempDir()
	repoPath := initScanJJRepo(t, root, "dirty")
	if err := os.WriteFile(filepath.Join(repoPath, "hello.txt"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write dirty file: %v", err)
	}

	scanOneJJRepo(t, root, config.OnlyUncommitted)
	scanOneJJRepo(t, root, config.OnlyDirty)
}

func TestGenerateScanReportFiltersJJUnpushedRepos(t *testing.T) {
	requireJJAndGit(t)

	repoPath := initScanTrackedJJRepo(t).WorkPath
	filePath := filepath.Join(repoPath, "README.md")
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open repo file: %v", err)
	}
	if _, err := f.WriteString("local\n"); err != nil {
		t.Fatalf("append local change: %v", err)
	}
	_ = f.Close()

	if err := exec.Command("jj", "-R", repoPath, "describe", "-m", "local change").Run(); err != nil {
		t.Fatalf("jj describe local change: %v", err)
	}
	if err := exec.Command("jj", "-R", repoPath, "bookmark", "move", "main", "-t", "@").Run(); err != nil {
		t.Fatalf("jj bookmark move main: %v", err)
	}
	if err := exec.Command("jj", "-R", repoPath, "new").Run(); err != nil {
		t.Fatalf("jj new: %v", err)
	}

	scanOneJJRepo(t, repoPath, config.OnlyUnpushed)
	scanOneJJRepo(t, repoPath, config.OnlyDirty)
}

func TestGenerateScanReportFiltersJJUnpulledRepos(t *testing.T) {
	requireJJAndGit(t)

	repo := initScanTrackedJJRepo(t)
	seedPath := repo.SeedPath
	workPath := repo.WorkPath

	workFile := filepath.Join(workPath, "README.md")
	f, err := os.OpenFile(workFile, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open work file: %v", err)
	}
	if _, err := f.WriteString("local\n"); err != nil {
		t.Fatalf("append local change: %v", err)
	}
	_ = f.Close()

	if err := exec.Command("jj", "-R", workPath, "describe", "-m", "local change").Run(); err != nil {
		t.Fatalf("jj describe local change: %v", err)
	}
	if err := exec.Command("jj", "-R", workPath, "bookmark", "move", "main", "-t", "@").Run(); err != nil {
		t.Fatalf("jj bookmark move main: %v", err)
	}
	if err := exec.Command("jj", "-R", workPath, "new").Run(); err != nil {
		t.Fatalf("jj new: %v", err)
	}

	seedFile := filepath.Join(seedPath, "README.md")
	f, err = os.OpenFile(seedFile, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open seed file: %v", err)
	}
	if _, err := f.WriteString("remote\n"); err != nil {
		t.Fatalf("append remote change: %v", err)
	}
	_ = f.Close()

	if err := exec.Command("git", "-C", seedPath, "add", "README.md").Run(); err != nil {
		t.Fatalf("git add remote: %v", err)
	}
	if err := exec.Command("git", "-C", seedPath, "commit", "-m", "remote change").Run(); err != nil {
		t.Fatalf("git commit remote: %v", err)
	}
	if err := exec.Command("git", "-C", seedPath, "push", "origin", "main").Run(); err != nil {
		t.Fatalf("git push remote: %v", err)
	}
	if err := exec.Command("jj", "-R", workPath, "git", "fetch").Run(); err != nil {
		t.Fatalf("jj git fetch: %v", err)
	}

	scanOneJJRepo(t, workPath, config.OnlyUnpulled)
	scanOneJJRepo(t, workPath, config.OnlyDirty)
}

func TestGenerateScanReportJSONIncludesJJFields(t *testing.T) {
	requireJJAndGit(t)

	repoPath := initScanTrackedJJRepo(t).WorkPath
	filePath := filepath.Join(repoPath, "README.md")
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open repo file: %v", err)
	}
	if _, err := f.WriteString("local\n"); err != nil {
		t.Fatalf("append local change: %v", err)
	}
	_ = f.Close()

	if err := exec.Command("jj", "-R", repoPath, "describe", "-m", "local change").Run(); err != nil {
		t.Fatalf("jj describe local change: %v", err)
	}
	if err := exec.Command("jj", "-R", repoPath, "bookmark", "move", "main", "-t", "@").Run(); err != nil {
		t.Fatalf("jj bookmark move main: %v", err)
	}
	if err := exec.Command("jj", "-R", repoPath, "new").Run(); err != nil {
		t.Fatalf("jj new: %v", err)
	}

	scanReport := GenerateScanReport(jjScanConfig(repoPath, config.OnlyAll))
	if len(scanReport.Warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", scanReport.Warnings)
	}
	if len(scanReport.RepoStates) != 1 {
		t.Fatalf("expected one repo, got %d: %#v", len(scanReport.RepoStates), scanReport.RepoStates)
	}

	data, err := json.Marshal(scanReport)
	if err != nil {
		t.Fatalf("marshal scan report: %v", err)
	}
	jsonReport := string(data)
	for _, field := range []string{`"vcsType":"jj"`, `"remoteStatus"`, `"outgoingCommits"`} {
		if !strings.Contains(jsonReport, field) {
			t.Fatalf("expected JSON report to include %s, got %s", field, jsonReport)
		}
	}
}
