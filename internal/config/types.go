package config

// Config holds all runtime options used by reposcan.
// Values may come from a config file and/or be overridden by CLI flags.
type Config struct {
	Roots     []string   `toml:"roots,omitempty"`
	DirIgnore []string   `toml:"dirignore,omitempty"`
	Only      OnlyFilter `toml:"only,omitempty"`

	Output Output `toml:"output"`

	// Max git checker workers
	MaxWorkers int `toml:"maxWorkers"`

	Version int `toml:"version"`
}

// Defaults returns a Config populated with sensible defaults suitable for
// typical local development machines.
func Defaults() Config {
	//home, err := os.UserHomeDir()

	var roots []string = nil
	roots = []string{"$HOME"}

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
		Type:     OutputTable,
		JSONPath: "",
	}

	return Config{
		Roots:      roots,
		DirIgnore:  defaultDirIgnore,
		Only:       OnlyDirty,
		Output:     newOutput,
		MaxWorkers: 8,
		Version:    1,
	}
}
