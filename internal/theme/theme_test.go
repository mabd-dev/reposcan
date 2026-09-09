package theme

import (
	"strings"
	"testing"
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

func TestLoadBase24Schema_NotFound(t *testing.T) {
	if _, err := LoadBase24Schema(SchemesDir + "no-such-scheme"); err == nil {
		t.Fatal("expected error for missing scheme")
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
