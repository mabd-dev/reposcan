package theme

import (
	"errors"
	"image/color"
	"io/fs"
	"reflect"
	"slices"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestLoadBase24Schema_Success(t *testing.T) {
	b, err := LoadBase24Schema(SchemesDir + "catppuccin-mocha")
	if err != nil {
		t.Fatalf("LoadBase24Schema: %v", err)
	}
	if b.Name == "" {
		t.Fatal("expected a scheme name")
	}
	if !strings.HasPrefix(b.Palette.Base00, "#") {
		t.Fatalf("expected hex palette color, got %q", b.Palette.Base00)
	}
}

func TestLoadBase24Schema_AppendsYAMLSuffix(t *testing.T) {
	// Passing a path without the .yaml extension must still load.
	b, err := LoadBase24Schema(SchemesDir + "catppuccin-mocha")
	if err != nil {
		t.Fatalf("LoadBase24Schema without extension: %v", err)
	}
	if b.Name == "" {
		t.Fatal("expected a scheme name")
	}
}

// A path that already ends in .yaml must not get a second suffix, and both
// forms must decode the same file.
func TestLoadBase24Schema_PathWithAndWithoutSuffix(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "without extension", path: SchemesDir + "catppuccin-mocha"},
		{name: "with extension", path: SchemesDir + "catppuccin-mocha.yaml"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := LoadBase24Schema(tt.path)
			if err != nil {
				t.Fatalf("LoadBase24Schema(%q): %v", tt.path, err)
			}
			if b.Name != "Catppuccin Mocha" || b.Variant != "dark" {
				t.Fatalf("name/variant = %q/%q, want %q/%q", b.Name, b.Variant, "Catppuccin Mocha", "dark")
			}
			if b.Palette.Base00 != "#1e1e2e" || b.Palette.Base17 != "#f5c2e7" {
				t.Fatalf("base00/base17 = %q/%q, want %q/%q", b.Palette.Base00, b.Palette.Base17, "#1e1e2e", "#f5c2e7")
			}
		})
	}
}

func TestLoadBase24Schema_NotFound(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "missing scheme", path: SchemesDir + "no-such-scheme"},
		{name: "empty scheme id", path: SchemesDir},
		{name: "wrong case", path: SchemesDir + "Catppuccin-Mocha"},
		{name: "missing schemes dir prefix", path: "catppuccin-mocha"},
		{name: "parent traversal", path: SchemesDir + "../" + SchemesDir + "catppuccin-mocha"},
		{name: "schemes dir itself", path: strings.TrimSuffix(SchemesDir, "/")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := LoadBase24Schema(tt.path)
			if !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("LoadBase24Schema(%q) error = %v, want fs.ErrNotExist", tt.path, err)
			}
			if b != (Base24Scheme{}) {
				t.Fatalf("expected zero scheme on error, got %+v", b)
			}
		})
	}
}

// Every scheme name exposed through Schemes (which config validation accepts)
// must match an embedded file and load into a fully colored ColorScheme;
// otherwise the TUI would silently fall back to the default scheme.
func TestSchemes_AllEmbeddedAndLoadable(t *testing.T) {
	files, err := fs.Glob(schemesFS, SchemesDir+"*.yaml")
	if err != nil {
		t.Fatalf("fs.Glob: %v", err)
	}
	embedded := make([]string, 0, len(files))
	for _, f := range files {
		embedded = append(embedded, strings.TrimSuffix(strings.TrimPrefix(f, SchemesDir), ".yaml"))
	}
	listed := slices.Clone(Schemes)
	slices.Sort(listed)
	slices.Sort(embedded)
	if !slices.Equal(listed, embedded) {
		t.Fatalf("Schemes is out of sync with embedded files:\n listed   %v\n embedded %v", listed, embedded)
	}

	for _, id := range Schemes {
		t.Run(id, func(t *testing.T) {
			c, err := LoadBase24(id)
			if err != nil {
				t.Fatalf("LoadBase24(%q): %v", id, err)
			}
			if c.ID != id {
				t.Fatalf("ID = %q, want %q", c.ID, id)
			}
			if c.Name == "" {
				t.Fatal("expected a scheme name")
			}
			for _, f := range colorFields(c) {
				if _, isNoColor := f.color.(lipgloss.NoColor); isNoColor || f.color == nil {
					t.Errorf("%s: palette value did not parse to a color (got %#v)", f.name, f.color)
				}
			}
		})
	}
}

type namedColor struct {
	name  string
	color color.Color
}

func colorFields(c ColorScheme) []namedColor {
	return []namedColor{
		{"Background", c.Background},
		{"Foreground", c.Foreground},
		{"Accent", c.Accent},
		{"Muted", c.Muted},
		{"Error", c.Error},
		{"Warning", c.Warning},
		{"Success", c.Success},
		{"Info", c.Info},
		{"Border", c.Border},
		{"BorderActive", c.BorderActive},
		{"TableHeader", c.TableHeader},
		{"TableRow", c.TableRow},
		{"TableAltRow", c.TableAltRow},
		{"PopupBackground", c.PopupBackground},
		{"PopupBorder", c.PopupBorder},
		{"PopupTitle", c.PopupTitle},
	}
}

// rgb builds the opaque color a "#rrggbb" palette entry parses to.
func rgb(hex uint32) color.RGBA {
	return color.RGBA{R: uint8(hex >> 16), G: uint8(hex >> 8), B: uint8(hex), A: 0xff}
}

func sameColor(a, b color.Color) bool {
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}

// Each ColorScheme field must come from its base24 palette slot. Expected
// values are the catppuccin-mocha palette entries noted next to each field.
func TestLoadBase24_MapsPaletteSlots(t *testing.T) {
	c, err := LoadBase24("catppuccin-mocha")
	if err != nil {
		t.Fatalf("LoadBase24: %v", err)
	}
	if c.Name != "Catppuccin Mocha" {
		t.Fatalf("Name = %q, want %q", c.Name, "Catppuccin Mocha")
	}

	want := map[string]color.RGBA{
		"Background":      rgb(0x1e1e2e), // base00
		"Foreground":      rgb(0xcdd6f4), // base05
		"Accent":          rgb(0x89b4fa), // base0D
		"Muted":           rgb(0x45475a), // base03
		"Error":           rgb(0xf38ba8), // base08
		"Warning":         rgb(0xfab387), // base09
		"Success":         rgb(0xa6e3a1), // base0B
		"Info":            rgb(0x94e2d5), // base0C
		"Border":          rgb(0x313244), // base02
		"BorderActive":    rgb(0x89b4fa), // base0D
		"TableHeader":     rgb(0x585b70), // base04
		"TableRow":        rgb(0xcdd6f4), // base05
		"TableAltRow":     rgb(0x181825), // base01
		"PopupBackground": rgb(0x1e1e2e), // base00
		"PopupBorder":     rgb(0x89b4fa), // base0D
		"PopupTitle":      rgb(0xf9e2af), // base0A
	}
	fields := colorFields(c)
	if len(fields) != len(want) {
		t.Fatalf("test table covers %d fields, colorFields returns %d", len(want), len(fields))
	}
	for _, f := range fields {
		t.Run(f.name, func(t *testing.T) {
			w, ok := want[f.name]
			if !ok {
				t.Fatalf("no expected color for field %s", f.name)
			}
			if f.color == nil || !sameColor(f.color, w) {
				t.Fatalf("%s = %#v, want %#v", f.name, f.color, w)
			}
		})
	}
}

func TestLoadBase24_Success(t *testing.T) {
	c, err := LoadBase24("catppuccin-mocha")
	if err != nil {
		t.Fatalf("LoadBase24: %v", err)
	}
	if c.ID != "catppuccin-mocha" {
		t.Fatalf("expected ID %q, got %q", "catppuccin-mocha", c.ID)
	}
	if c.Name == "" {
		t.Fatal("expected populated name")
	}
	// Spot-check that palette entries were mapped into the color scheme.
	if c.Background == nil || c.Foreground == nil || c.Accent == nil {
		t.Fatal("expected backend/foreground/accent colors to be mapped")
	}
}

func TestLoadBase24_PropagatesLoadError(t *testing.T) {
	if _, err := LoadBase24("no-such-scheme"); err == nil {
		t.Fatal("expected error to propagate from LoadBase24Schema")
	}
}

func TestCreateColors_Success(t *testing.T) {
	c, err := CreateColors("catppuccin-mocha")
	if err != nil {
		t.Fatalf("CreateColors: %v", err)
	}
	if c.ID != "catppuccin-mocha" {
		t.Fatalf("expected requested scheme, got %q", c.ID)
	}
}

func TestCreateColors_FallsBackToDefaultOnMissingScheme(t *testing.T) {
	c, err := CreateColors("definitely-not-a-real-scheme")
	if err != nil {
		t.Fatalf("expected fallback to default, got error: %v", err)
	}
	if c.ID != defaultSchemeID {
		t.Fatalf("expected fallback to default scheme %q, got %q", defaultSchemeID, c.ID)
	}
}

// catppuccin-mocha is also the default, so a non-default scheme is needed to
// tell a successful load apart from the fallback.
func TestCreateColors_SelectsSchemeOrDefault(t *testing.T) {
	def, err := LoadBase24(defaultSchemeID)
	if err != nil {
		t.Fatalf("LoadBase24(default): %v", err)
	}
	latte, err := LoadBase24("catppuccin-latte")
	if err != nil {
		t.Fatalf("LoadBase24(catppuccin-latte): %v", err)
	}

	tests := []struct {
		name     string
		schemeID string
		want     ColorScheme
	}{
		{name: "non-default scheme", schemeID: "catppuccin-latte", want: latte},
		{name: "default scheme", schemeID: defaultSchemeID, want: def},
		{name: "unknown scheme", schemeID: "definitely-not-a-real-scheme", want: def},
		{name: "empty scheme id", schemeID: "", want: def},
		{name: "path traversal", schemeID: "../" + SchemesDir + "catppuccin-latte", want: def},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateColors(tt.schemeID)
			if err != nil {
				t.Fatalf("CreateColors(%q): %v", tt.schemeID, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("CreateColors(%q) = %s (%s), want %s (%s)", tt.schemeID, got.ID, got.Name, tt.want.ID, tt.want.Name)
			}
		})
	}
}

func TestCreateStyles_PopulatesExpectedStyles(t *testing.T) {
	c, err := LoadBase24("catppuccin-mocha")
	if err != nil {
		t.Fatalf("LoadBase24: %v", err)
	}
	styles := CreateStyles(c)

	// Styles share the same underlying type; ensure the ones we depend on are
	// not the zero value.
	if styles.Base.Render("x") == "" {
		t.Fatal("expected Base style to render content")
	}
	if styled := styles.TableHeader.Render("header"); styled == "" || !strings.Contains(styled, "header") {
		t.Fatalf("expected TableHeader style to render text, got %q", styled)
	}
}

// lipgloss.Style is not comparable, so BoxFor is verified through the rendered
// output: the active box carries the active border color and the muted box the
// muted border color.
func TestStyles_BoxFor(t *testing.T) {
	c, _ := LoadBase24("catppuccin-mocha")
	styles := CreateStyles(c)

	if got := styles.BoxFor(true).Render("x"); got != styles.Box.Render("x") {
		t.Fatalf("BoxFor(true) mismatch:\n got %q\n want %q", got, styles.Box.Render("x"))
	}
	if got := styles.BoxFor(false).Render("x"); got != styles.BoxMuted.Render("x") {
		t.Fatalf("BoxFor(false) mismatch:\n got %q\n want %q", got, styles.BoxMuted.Render("x"))
	}
}

func TestCreateStyles_BoxBorders(t *testing.T) {
	c, _ := LoadBase24("catppuccin-mocha")
	styles := CreateStyles(c)

	// Box and BoxMuted use a rounded border; confirm the rendered border
	// contains rounded characters (╭) rather than being unstyled.
	if got := styles.Box.Render("x"); !strings.Contains(got, "╭") {
		t.Fatalf("expected rounded border on Box, got %q", got)
	}
	if got := styles.BoxMuted.Render("x"); !strings.Contains(got, "╭") {
		t.Fatalf("expected rounded border on BoxMuted, got %q", got)
	}
}
