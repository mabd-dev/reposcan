package theme

import (
	"embed"
	"errors"
	"io/fs"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

// Every embedded scheme decodes cleanly, so the decode-error path and the
// "default scheme also missing" path are reached by swapping the package-level
// schemesFS and defaultSchemeID for the duration of a test. Tests in this file
// must not call t.Parallel.

//go:embed testdata/malformed.yaml
var malformedSchemeFS embed.FS

func withSchemesFS(t *testing.T, fsys embed.FS) {
	t.Helper()
	orig := schemesFS
	schemesFS = fsys
	t.Cleanup(func() { schemesFS = orig })
}

func withDefaultScheme(t *testing.T, id string) {
	t.Helper()
	orig := defaultSchemeID
	defaultSchemeID = id
	t.Cleanup(func() { defaultSchemeID = orig })
}

func TestLoadBase24Schema_MalformedYAML(t *testing.T) {
	withSchemesFS(t, malformedSchemeFS)

	b, err := LoadBase24Schema("testdata/malformed")
	var typeErr *yaml.TypeError
	if !errors.As(err, &typeErr) {
		t.Fatalf("LoadBase24Schema error = %v, want *yaml.TypeError", err)
	}
	// yaml.v3 fills the fields it could decode (name, variant) before
	// reporting the palette error; none of that may reach the caller.
	if b != (Base24Scheme{}) {
		t.Fatalf("expected zero scheme on decode error, got %+v", b)
	}
}

func TestCreateColors_DefaultScheme(t *testing.T) {
	latte, err := LoadBase24("catppuccin-latte")
	if err != nil {
		t.Fatalf("LoadBase24(catppuccin-latte): %v", err)
	}

	tests := []struct {
		name          string
		defaultID     string
		schemeID      string
		want          ColorScheme
		wantErrorPath string // non-empty: expect fs.ErrNotExist for this path
	}{
		{
			name:      "fallback uses the configured default",
			defaultID: "catppuccin-latte",
			schemeID:  "definitely-not-a-real-scheme",
			want:      latte,
		},
		{
			name:      "valid scheme does not need the default",
			defaultID: "no-such-default",
			schemeID:  "catppuccin-latte",
			want:      latte,
		},
		{
			name:          "missing scheme and missing default",
			defaultID:     "no-such-default",
			schemeID:      "definitely-not-a-real-scheme",
			want:          ColorScheme{},
			wantErrorPath: SchemesDir + "no-such-default.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withDefaultScheme(t, tt.defaultID)

			got, err := CreateColors(tt.schemeID)
			if tt.wantErrorPath == "" {
				if err != nil {
					t.Fatalf("CreateColors(%q): %v", tt.schemeID, err)
				}
			} else {
				var pathErr *fs.PathError
				if !errors.Is(err, fs.ErrNotExist) || !errors.As(err, &pathErr) {
					t.Fatalf("CreateColors(%q) error = %v, want fs.ErrNotExist *fs.PathError", tt.schemeID, err)
				}
				if pathErr.Path != tt.wantErrorPath {
					t.Fatalf("error path = %q, want %q", pathErr.Path, tt.wantErrorPath)
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("CreateColors(%q) = %+v, want %+v", tt.schemeID, got, tt.want)
			}
		})
	}
}
