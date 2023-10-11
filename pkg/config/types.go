package config

type Config struct {
	// TmuxBaseIndex is the base index (base-index in your tmux.conf)
	TMUXBaseIndex int `yaml:"tmux-base-index"`

	// ProjectProvider is the configuration for projects
	ProjectProvider *ProjectProviderConfig `yaml:"project"`

	// SSHProvider is the configuration for ssh connections
	SSHProvider *SSHProviderConfig `yaml:"ssh"`

	WindowTemplates map[string][]Window `yaml:"window-template"`
}

type ProjectProviderConfig struct {
	Enabled bool `yaml:"enabled"`

	// Sources is a list of source directories that should be scanned
	SourceDirectories []SourceDirectory `yaml:"directories"`

	// Checks is a list of files or directories that should be checked, e.g. ".git", ".gitignore"
	Checks []string `yaml:"checks"`

	// DisplayFormat is the format that should be used to display the project name
	DisplayFormat ProjectDisplayFormat `yaml:"display-format"`
}

type SourceDirectory struct {
	// Directory is the absolute path to the source directory
	Directory string `yaml:"path"`

	// Depth is the maximum depth of subdirectories that should be scanned
	Depth int `yaml:"depth"`

	// Exclude is a list of directories that should be excluded from the scan
	Exclude []string `yaml:"exclude"`

	// Tags can be used to filter directories
	Tags []string `yaml:"tags"`
}

type SSHProviderConfig struct {
	Enabled bool `yaml:"enabled"`

	// Mode controls how sessions or windows are created for SSH connections
	Mode SSHMode `yaml:"mode"`
}

type Window struct {
	// Name of the window
	Name string `yaml:"name"`

	// Commands that should be executed in the window
	Commands []string `yaml:"commands,omitempty"`

	// Default indicates if this window should be selected by default
	Default bool `yaml:"default,omitempty"`
}

type ProjectDisplayFormat string

const (
	AbsolutePath ProjectDisplayFormat = "abs"
	RelativePath ProjectDisplayFormat = "rel"
	BaseName     ProjectDisplayFormat = "base"
)

type SSHMode string

const (
	SSHSessionMode SSHMode = "session"
	SSHWindowMode  SSHMode = "window"
)
