package config

type OnlyFilter string

const (
	OnlyAll OnlyFilter = "all"

	// Have uncommited files
	OnlyUncommited = "uncommited"
)

func (f OnlyFilter) IsValid() bool {
	switch f {
	case OnlyAll, OnlyUncommited:
		return true
	}
	return false
}

type Config struct {
	Roots     []string   `json:"roots,omitempty"`
	DirIgnore []string   `json:"dirignore,omitempty"`
	Only      OnlyFilter `json:"only,omitempty"`

	// Write report json to path, ignored if empty
	JsonOutputPath string `json:"jsonOutputPath,omitempty"`

	// Print json on std out,
	JsonStdOut bool `json:"jsonStdOut"`

	Version int `json:"version"`
}

func Defaults() Config {
	return Config{
		Roots:          nil,
		DirIgnore:      nil,
		Only:           OnlyAll,
		JsonOutputPath: "",
		JsonStdOut:     false,
		Version:        1,
	}
}
