package config

const (
	DefaultConfigDir  = "/.config/reposcan/"
	DefaultConfigToml = "config.toml"
)

type Paths struct {
	ConfigDir      string
	ConfigFilePath string
}

func DefaultPaths() Paths {
	return Paths{
		ConfigDir:      DefaultConfigDir,
		ConfigFilePath: DefaultConfigDir + DefaultConfigToml,
	}
}
