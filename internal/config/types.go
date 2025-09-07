package config

type Config struct {
	Roots     []string   `json:"roots,omitempty"`
	DirIgnore []string   `json:"dirignore,omitempty"`
	Only      OnlyFilter `json:"only,omitempty"`

	// Write report json to path, ignored if empty
	JsonOutputPath string `json:"jsonOutputPath,omitempty"`

	// Print json on std out,
	Output OutputFormat `json:"output"`

	// Max git checker workers
	MaxWorkers int `json:"maxWorkers"`

	Version int `json:"version"`
}

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

	return Config{
		Roots:          roots,
		DirIgnore:      defaultDirIgnore,
		Only:           OnlyDirty,
		JsonOutputPath: "",
		Output:         OutputTable,
		MaxWorkers:     8,
		Version:        1,
	}
}
