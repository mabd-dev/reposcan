package config

import "os"

// Config holds all runtime options used by reposcan.
// Values may come from a config file and/or be overridden by CLI flags.
type Config struct {
	Roots     []string   `toml:"roots,omitempty"`
	DirIgnore []string   `toml:"dirignore,omitempty"`
	Only      OnlyFilter `toml:"only,omitempty"`

	Output Output `toml:"output"`

	Tui Tui `toml:"tui"`

	// CountStashAsDirty, when true, makes repos with only stashed work count as
	// dirty for IsDirty-based filtering (--filter dirty) and the dirty total.
	// The --filter stash value is unaffected by this setting.
	CountStashAsDirty bool `toml:"countStashAsDirty,omitempty"`

	// Max git checker workers
	MaxWorkers int `toml:"maxWorkers"`

	// Debug if true, enable logging to a file in [DefaultLogFileDir]
	Debug       bool `toml:"debug"`
	NoTelemetry bool `toml:"no-telemetry"`

	Version int `toml:"version"`
}

// Tui holds display options for the interactive renderer.
type Tui struct {
	ShowVCS *bool `toml:"showVCS,omitempty"`
}

// ShowVCSColumn reports whether the interactive table should render the VCS
// column. An unset value defaults to true for compatibility with existing
// config files.
func (c Config) ShowVCSColumn() bool {
	return c.Tui.ShowVCS == nil || *c.Tui.ShowVCS
}

// Defaults returns a Config populated with sensible defaults suitable for
// typical local development machines.
func Defaults() Config {
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}

	var roots []string = nil
	roots = []string{home}

	defaultDirIgnore := []string{
		// --- Package managers / deps ---
		"**/node_modules/**",
		"**/vendor/**",
		"**/.venv/**",
		"**/venv/**",
		"**/.m2/**",
		"**/.gradle/**",
		"**/.cargo/**",
		"**/.gradle/**",
		"**/.kotlin/**",
		"**/.java/**",
		"**/.cargo/**",
		"**/.zen/**",
		"**/.bun/**",
		"**/.codex/**",
		"**/.android/**",
		"**/.config/Google/**",
		"**/.config/JetBrains/**",
		"**/target/**",

		// --- Build / dist ---
		"**/build/**",
		"**/dist/**",
		"**/.next/**",
		"**/.nuxt/**",

		// --- Cache & temp ---
		"**/.cache/**",
		"**/.local/**",
		"**/.pytest_cache/**",

		// --- IDE / tooling ---
		"**/.idea/**",
		"**/.vscode/**",
		"**/.terraform/**",
		"**/.docker/**",

		// --- OS metadata ---
		"**/.DS_Store", // macOS
		"**/Thumbs.db", // Windows

		// --- Linux system dirs ---
		"/proc/**",
		"/sys/**",
		"/dev/**",
		"/run/**",
		"/tmp/**",
		"/var/log/**",
		"/var/tmp/**",

		// --- macOS system dirs ---
		"/System/**",
		"/Library/**",
		"~/Library/**",
	}

	newOutput := Output{
		Type:     OutputInteractive,
		JSONPath: "",
	}

	showVCS := true

	return Config{
		Roots:      roots,
		DirIgnore:  defaultDirIgnore,
		Only:       OnlyDirty,
		Output:     newOutput,
		Tui:        Tui{ShowVCS: &showVCS},
		MaxWorkers: 8,
		Debug:      false,
		Version:    1,
	}
}
