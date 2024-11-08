package config

type FinderConfig struct {
	// Executable is the fuzzy finder, e.g. "fzf" or "embedded"
	Executable string `yaml:"executable"`

	// Preview indicates if the preview should be shown
	Preview bool `yaml:"preview"`

	// FZFPreview can be used to overwrite the option delimiter
	FZFDelimiter string `yaml:"fzf-delimiter"`
}

type Layout struct {
	// Apps contains the list of apps that should be started
	Apps []App `yaml:"apps"`

	// Rules is a list of rules, at least one must match for this layout to be selected
	Rules []string `yaml:"rules,omitempty"`

	// ClearWorkspace indicates if the workspace should be cleared before starting the applications (only applies to window managers, default: false)
	ClearWorkspace bool `yaml:"clear-workspace,omitempty"`
}

type App struct {
	// Name of the window
	Name string `yaml:"name"`

	// Commands that should be executed in the window
	Commands []Command `yaml:"commands,omitempty"`

	// Default indicates if this window should be selected by default
	Default bool `yaml:"default,omitempty"`

	// Rules is a list of rules, at least one must match for the window to be created
	Rules []string `yaml:"rules,omitempty"`

	// GUI indicates that this app is a GUI application (will not be started in a terminal)
	GUI bool `yaml:"gui,omitempty"`

	// Group a app belongs to, only the first matching option within a group will be used
	Group string `yaml:"group,omitempty"`
}

type Command struct {
	// Command that should be executed
	Command string `yaml:"command"`

	// Rules is a list of rules, at least one must match for the window to be created
	Rules []string `yaml:"rules,omitempty"`
}
