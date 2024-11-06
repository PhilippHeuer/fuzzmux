package config

type FinderConfig struct {
	// Executable is the fuzzy finder, e.g. "fzf" or "embedded"
	Executable string `yaml:"executable"`

	// Preview indicates if the preview should be shown
	Preview bool `yaml:"preview"`

	// FZFPreview can be used to overwrite the option delimiter
	FZFDelimiter string `yaml:"fzf-delimiter"`
}

type ProjectModuleConfig struct {
	// Name is used to override the default module name
	Name string `yaml:"name,omitempty"`

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

type SSHModuleConfig struct {
	// Name is used to override the default module name
	Name string `yaml:"name,omitempty"`

	// ConfigFile is used in case your ssh config is not in the default location
	ConfigFile string `yaml:"file"`

	// StartDirectory is used to define the current working directory, supports template variables
	StartDirectory string `yaml:"start-directory"`

	// Mode controls how sessions or windows are created for SSH connections
	Mode SSHMode `yaml:"mode"`
}

type KubernetesModuleConfig struct {
	// Name is used to override the default module name
	Name string `yaml:"name,omitempty"`

	// Clusters is a list of kubernetes clusters that should be scanned
	Clusters []KubernetesCluster `yaml:"clusters"`

	// StartDirectory is used to define the current working directory, supports template variables
	StartDirectory string `yaml:"start-directory"`
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

type USQLModuleConfig struct {
	// Name is used to override the default module name
	Name string `yaml:"name,omitempty"`

	// ConfigFile is used in case your usql config is not in the default location
	ConfigFile string `yaml:"file"`

	// StartDirectory is used to define the current working directory, supports template variables
	StartDirectory string `yaml:"start-directory"`
}

type StaticModuleConfig struct {
	// Name is used to override the default module name
	Name string `yaml:"name,omitempty"`

	// Options is a list of static options
	StaticOptions []StaticOption `yaml:"options"`
}

type LDAPModuleConfig struct {
	// Name is used to override the default module name
	Name string `yaml:"name,omitempty"`

	// Host is the LDAP server hostname or IP address
	Host string `yaml:"host"`

	// BaseDistinguishedName (DN) for LDAP base search (e.g., "dc=example,dc=com")
	BaseDistinguishedName string `yaml:"base-dn"`

	// BindDistinguishedName (DN) used for LDAP binding (e.g., "cn=admin,dc=example,dc=com")
	BindDistinguishedName string `yaml:"bind-dn"`

	// Password for LDAP bind user
	BindPassword string `yaml:"bind-password"`

	// Filter is the LDAP search filter (e.g., "(&(objectClass=organizationalPerson))")
	Filter string `yaml:"filter"`
}

type KeycloakModuleConfig struct {
	// Name is used to override the default module name
	Name string `yaml:"name,omitempty"`

	// Host is the Keycloak server hostname or IP address
	Host string `yaml:"host"`

	// RealmName is the Keycloak realm name
	RealmName string `yaml:"realm"`

	// Username is the Keycloak admin username
	Username string `yaml:"username"`

	// Password is the Keycloak admin password
	Password string `yaml:"password"`

	// Query is a list of content types that should be queried
	Query []KeycloakContent `yaml:"query"`
}

type KeycloakContent string

const (
	KeycloakUser   KeycloakContent = "user"
	KeycloakClient KeycloakContent = "client"
	KeycloakGroup  KeycloakContent = "group"
)

type BackstageModuleConfig struct {
	// Name is used to override the default module name
	Name string `yaml:"name,omitempty"`

	// Host is the Backstage hostname or IP address
	Host string `yaml:"host"`

	// Query is a list of content types that should be queried
	Query []string `yaml:"query"`
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
