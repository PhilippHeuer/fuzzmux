package types

type FieldMapping struct {
	// Source is the source field name
	Source string `yaml:"source"`

	// Format optionally specifies a format of the source field, e.g. "unixtsmillis", "ldaptime", ...
	Format string `yaml:"format"`

	// Target is the target field name
	Target string `yaml:"target"`
}
