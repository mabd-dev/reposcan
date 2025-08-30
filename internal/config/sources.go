package config

// This file just documents the intended precedence; you’ll implement it later.
//
// Precedence (lowest → highest):
//   1) Defaults()                                 (in-code defaults)
//   2) Config file (JSON/TOML)                     (~/.config/gitchecker/config.toml)
//   3) Environment variables                       (e.g., GITCHECKER_CONFIG)
//   4) CLI flags                                   (--root, --json, --only, ...)
//
// Keep this comment as a contract for future implementation.

// Source is a descriptor you might use later when merging; not required today.
type Source string

const (
	SourceDefaults Source = "defaults"
	SourceFile     Source = "file"
	SourceEnv      Source = "env"
	SourceFlags    Source = "flags"
)
