package reposcan

import (
	"fmt"
	"testing"

	"github.com/mabd-dev/reposcan/internal"
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
