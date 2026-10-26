package reposcan

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mabd-dev/reposcan/internal"
	"github.com/spf13/cobra"
)

func TestGetCurrentVersion(t *testing.T) {
	expected := internal.VERSION

	if currentVersion := getCurrentVersion(); currentVersion != expected {
		t.Fatalf("expected current version =%v, found %v\n", expected, currentVersion)
	}
}

func TestNeedToUpdate(t *testing.T) {
	tests := []struct {
		currentVersion string
		latestVersion  string
		expectedOutput bool
		expectedError  bool
	}{
		// current > latest
		{
			currentVersion: "1.0.0",
			latestVersion:  "0.9.0",
			expectedOutput: false,
		},
		// current < latest
		{
			currentVersion: "1.0.0",
			latestVersion:  "1.1.0",
			expectedOutput: true,
		},
		{
			currentVersion: "1.1.0",
			latestVersion:  "1.1.1",
			expectedOutput: true,
		},
		// current == latest
		{
			currentVersion: "1.1.0",
			latestVersion:  "1.1.0",
			expectedOutput: false,
		},
		{
			currentVersion: "1.1.2",
			latestVersion:  "1.1.2",
			expectedOutput: false,
		},
		// currentVersion is not a valid
		{
			currentVersion: "1.1.",
			latestVersion:  "1.0.0",
			expectedOutput: false,
			expectedError:  true,
		},
		{
			currentVersion: "1.1.x",
			latestVersion:  "1.0.0",
			expectedOutput: false,
			expectedError:  true,
		},
		// latestVersion is not a valid
		{
			currentVersion: "1.0.0",
			latestVersion:  "1.0.",
			expectedOutput: false,
			expectedError:  true,
		},
		{
			currentVersion: "1.0.0",
			latestVersion:  "1.0.x",
			expectedOutput: false,
			expectedError:  true,
		},
		// latestVersion is alpha
		{
			currentVersion: "1.0.0",
			latestVersion:  "1.0.0-alpha",
			expectedOutput: false,
		},
		{
			currentVersion: "1.0.0",
			latestVersion:  "1.0.0-beta",
			expectedOutput: false,
		},
	}

	for i, test := range tests {
		testName := fmt.Sprintf("needToUpdate(%v)", i)
		t.Run(testName, func(t *testing.T) {
			needToUpdate, err := needToUpdate(test.currentVersion, test.latestVersion)

			if err == nil && test.expectedError {
				t.Fatal("expecting error but got nothing")
			}

			if err != nil && !test.expectedError {
				t.Fatalf("not expecting error but found erro=%v", err.Error())
			}

			if needToUpdate != test.expectedOutput {
				t.Fatalf("currentVersion=%v, latestVersion%v, expectedOutput=%v, found=%v", test.currentVersion, test.latestVersion, test.expectedOutput, needToUpdate)
			}
		})
	}
}

func TestUpdateFunction(t *testing.T) {
	fakeShellRunner := func(name string, args ...string) error {
		var sb strings.Builder

		sb.WriteString(name)
		for _, arg := range args {
			sb.WriteString(" ")
			sb.WriteString(arg)
		}
		return fmt.Errorf("%v", sb.String())
	}

	alias := "fake-reposcan"
	err := update(alias, fakeShellRunner)
	if err == nil {
		t.Fatal("expected error, found nil")
	}

	msg := err.Error()
	expectedMsg := fmt.Sprintf("sh -c curl -fsSL https://raw.githubusercontent.com/mabd-dev/reposcan/main/install.sh | ALIAS=%v sh", alias)

	if msg != expectedMsg {
		t.Fatalf("wrong cmd, expected=%v, found=%v", expectedMsg, msg)
	}
}

func TestUpdateFunction_InvalidAlias(t *testing.T) {
	invalidAliases := []string{
		"", "-", "repo/scan", "repo scan", "rs;whoami", "-reposcan",
		"repo'scan", `repo"scan`, `repo\scan`,
	}

	fakeShellRunner := func(name string, args ...string) error {
		return nil
	}

	for _, alias := range invalidAliases {
		err := update(alias, fakeShellRunner)
		if err == nil {
			t.Fatalf("expected error, found nil, alias=%v", alias)
		}
	}
}

// serveRelease points latestReleaseURL at a test server for the test's lifetime.
func serveRelease(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	old := latestReleaseURL
	latestReleaseURL = srv.URL
	t.Cleanup(func() { latestReleaseURL = old })
	return srv
}

func TestGetLatestVersion(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		want    string
		wantErr bool
	}{
		{name: "returns the release tag", status: http.StatusOK, body: `{"tag_name":"v9.9.9"}`, want: "v9.9.9"},
		{name: "rejects a non-200 status", status: http.StatusForbidden, body: `{}`, wantErr: true},
		{name: "rejects a malformed body", status: http.StatusOK, body: `not json`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serveRelease(t, tt.status, tt.body)

			got, err := getLatestVersion()
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetLatestVersion_Unreachable(t *testing.T) {
	serveRelease(t, http.StatusOK, "").Close()

	if _, err := getLatestVersion(); err == nil {
		t.Fatal("expected a connection error")
	}
}

func TestUpdateCmdRunE(t *testing.T) {
	newer := `{"tag_name":"v999.0.0"}`

	tests := []struct {
		name      string
		status    int
		body      string
		aliasFlag string // empty means the command has no alias flag
		runnerErr error
		wantAlias string // empty means the installer must not run
		wantErr   bool
	}{
		{
			name:   "skips install when already on the latest version",
			status: http.StatusOK, body: fmt.Sprintf(`{"tag_name":%q}`, internal.VERSION),
		},
		{
			name:   "installs under the requested alias",
			status: http.StatusOK, body: newer, aliasFlag: "rs", wantAlias: "rs",
		},
		{
			name:   "falls back to the default alias when the flag is missing",
			status: http.StatusOK, body: newer, wantAlias: "reposcan",
		},
		{
			name:   "reports an installer failure",
			status: http.StatusOK, body: newer, aliasFlag: "rs", runnerErr: errors.New("boom"),
			wantAlias: "rs", wantErr: true,
		},
		{
			name:   "reports a release lookup failure",
			status: http.StatusInternalServerError, body: `{}`, wantErr: true,
		},
		{
			name:   "reports an unparseable release tag",
			status: http.StatusOK, body: `{"tag_name":"not-a-version"}`, wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serveRelease(t, tt.status, tt.body)

			var script string
			old := realShellRunner
			realShellRunner = func(name string, args ...string) error {
				script = args[len(args)-1]
				return tt.runnerErr
			}
			t.Cleanup(func() { realShellRunner = old })

			cmd := &cobra.Command{Use: "update"}
			if tt.aliasFlag != "" {
				cmd.Flags().String("alias", "reposcan", "")
				parseFlags(t, cmd, "--alias", tt.aliasFlag)
			}

			var err error
			captureStdout(t, func() { err = updateCmd.RunE(cmd, nil) })

			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantAlias == "" && script != "" {
				t.Fatalf("installer ran unexpectedly: %q", script)
			}
			if tt.wantAlias != "" && !strings.Contains(script, "ALIAS="+tt.wantAlias+" sh") {
				t.Fatalf("installer script %q does not use alias %q", script, tt.wantAlias)
			}
		})
	}
}

func TestRealShellRunner(t *testing.T) {
	// Re-running the test binary with no tests selected is a portable command
	// that exits 0; a path that does not exist must surface an error.
	var err error
	captureStdout(t, func() { err = realShellRunner(os.Args[0], "-test.list=^$") })
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if err := realShellRunner(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("expected an error for a missing executable")
	}
}
