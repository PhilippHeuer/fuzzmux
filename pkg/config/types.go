package config

type Config struct {
	// ProjectProvider is the configuration for projects
	ProjectProvider *ProjectProviderConfig `yaml:"project"`

	// SSHProvider is the configuration for ssh connections
	SSHProvider *SSHProviderConfig `yaml:"ssh"`

	// KubernetesProvider is the configuration for k8s connections
	KubernetesProvider *KubernetesProviderConfig `yaml:"kubernetes"`

	// StaticProvider allows to define static options
	StaticProvider *StaticProviderConfig `yaml:"static"`

	// Layouts is a map of tmux layouts
	Layouts map[string]Layout `yaml:"layouts"`

	// Finder
	Finder *FinderConfig `yaml:"finder"`
}

type FinderConfig struct {
	// Executable is the fuzzy finder, e.g. "fzf" or "embedded"
	Executable string `yaml:"executable"`

	// Preview indicates if the preview should be shown
	Preview bool `yaml:"preview"`

	// FZFPreview can be used to overwrite the option delimiter
	FZFDelimiter string `yaml:"fzf-delimiter"`
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

type KubernetesProviderConfig struct {
	Enabled bool `yaml:"enabled"`

	// Clusters is a list of kubernetes clusters that should be scanned
	Clusters []KubernetesCluster `yaml:"clusters"`
}

type KubernetesCluster struct {
	// Name of the cluster
	Name string `yaml:"name"`

	// Tags that apply to the cluster
	Tags []string `yaml:"tags"`

	// OpenShift indicates if this is an OpenShift cluster (default: false)
	OpenShift bool `yaml:"openshift"`

	// KubeConfig is the absolute path to the kubeconfig file
	KubeConfig string `yaml:"kubeconfig"`
}

type StaticProviderConfig struct {
	Enabled bool `yaml:"enabled"`

	// Options is a list of static options
	StaticOptions []StaticOption `yaml:"options"`
}

type StaticOption struct {
	// Id is a unique identifier for the option
	Id string `yaml:"id"`

	// DisplayName is the name that should be displayed in the fuzzy finder
	DisplayName string `yaml:"display-name"`

	// Name is the name of the option
	Name string `yaml:"name"`

	// StartDirectory is the initial working directory
	StartDirectory string `yaml:"start-directory"`

	// Tags can be used to filter options
	Tags []string `yaml:"tags"`

	// Context
	Context map[string]string `yaml:"context"`

	// Layout can be used to override the default layout used by the option (e.g. ssh/project)
	Layout string `yaml:"layout"`

	// Preview to render in the preview window
	Preview string `yaml:"preview"`
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

type ProjectDisplayFormat string

const (
	AbsolutePath ProjectDisplayFormat = "absolute"
	RelativePath ProjectDisplayFormat = "relative"
	BaseName     ProjectDisplayFormat = "base"
)

type SSHMode string

const (
	SSHSessionMode SSHMode = "session"
	SSHWindowMode  SSHMode = "window"
)
