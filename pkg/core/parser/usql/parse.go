package usql

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Connections map[string]Connection `yaml:"connections"`
}

type Connection struct {
	Protocol string `yaml:"protocol"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Hostname string `yaml:"hostname"`
	Port     int    `yaml:"port"`
	Instance string `yaml:"instance"`
	Database string `yaml:"database"`
}

func (c *Connection) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	if err := unmarshal(&str); err == nil {
		return c.parseFromString(str)
	}

	var aux struct {
		Protocol string `yaml:"protocol"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Hostname string `yaml:"hostname"`
		Port     int    `yaml:"port"`
		Instance string `yaml:"instance"`
		Database string `yaml:"database"`
	}
	if err := unmarshal(&aux); err != nil {
		return err
	}

	*c = aux
	return nil
}

func (c *Connection) parseFromString(str string) error {
	parts := strings.SplitN(str, "://", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid connection string: %s", str)
	}

	c.Protocol = parts[0]

	credentialsAndHost := strings.SplitN(parts[1], "@", 2)
	if len(credentialsAndHost) != 2 {
		return fmt.Errorf("invalid connection string: %s", str)
	}

	credentials := strings.SplitN(credentialsAndHost[0], ":", 2)
	if len(credentials) != 2 {
		return fmt.Errorf("invalid connection string: %s", str)
	}

	c.Username = credentials[0]
	c.Password = credentials[1]

	hostAndPort := strings.SplitN(credentialsAndHost[1], ":", 2)
	c.Hostname = hostAndPort[0]

	if len(hostAndPort) == 2 {
		_, err := fmt.Sscanf(hostAndPort[1], "%d", &c.Port)
		if err != nil {
			return fmt.Errorf("invalid connection string: %s", str)
		}
	}

	return nil
}

// ParseFile parses the usql config file
func ParseFile(path string) (*Config, error) {
	// read file
	configBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read usql config file: %w", err)
	}

	// parse file
	var config Config
	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse usql config file: %w", err)
	}

	return &config, nil
}
