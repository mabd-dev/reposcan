package config

import "os"

// Config holds all runtime options used by reposcan.
// Values may come from a config file and/or be overridden by CLI flags.
type Config struct {
	Roots     []string   `toml:"roots,omitempty"`
	DirIgnore []string   `toml:"dirignore,omitempty"`
	Only      OnlyFilter `toml:"only,omitempty"`

	Output Output `toml:"output"`

	// CountStashAsDirty, when true, makes repos with only stashed work count as
	// dirty for IsDirty-based filtering (--filter dirty) and the dirty total.
	// The --filter stash value is unaffected by this setting.
	CountStashAsDirty bool `toml:"count_stash_as_dirty,omitempty"`

	// Max git checker workers
	MaxWorkers int `toml:"maxWorkers"`

	// Debug if true, enable logging to a file in [DefaultLogFileDir]
	Debug       bool `toml:"debug"`
	NoTelemetry bool `toml:"no-telemetry"`

	Version int `toml:"version"`
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

	return Config{
		Roots:      roots,
		DirIgnore:  defaultDirIgnore,
		Only:       OnlyDirty,
		Output:     newOutput,
		MaxWorkers: 8,
		Debug:      false,
		Version:    1,
	}
}
