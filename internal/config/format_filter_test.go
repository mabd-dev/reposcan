package config

import "testing"

func TestCreateOutputFormat_ValidValues(t *testing.T) {
    cases := []string{"json", "table", "none", "JSON", "Table"}
    for _, c := range cases {
        if _, err := CreateOutputFormat(c); err != nil {
            t.Fatalf("expected valid output format for %q, got error: %v", c, err)
        }
    }
}

func TestCreateOutputFormat_InvalidValue(t *testing.T) {
    if _, err := CreateOutputFormat("yaml"); err == nil {
        t.Fatalf("expected error for invalid output format")
    }
}

func TestCreateOnlyFilter_ValidValues(t *testing.T) {
    cases := []string{"all", "dirty", "uncommitted", "unpushed", "unpulled", "ALL", "Dirty", "Uncommitted"}
    for _, c := range cases {
        if _, err := CreateOnlyFilter(c); err != nil {
            t.Fatalf("expected valid only filter for %q, got error: %v", c, err)
        }
    }
}

func TestCreateOnlyFilter_InvalidValue(t *testing.T) {
    if _, err := CreateOnlyFilter("invalid"); err == nil {
        t.Fatalf("expected error for invalid only filter")
    }
}
