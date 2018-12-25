package mono

import (
	"errors"
	"path/filepath"
	"strings"
)

const (
	defaultBinaryName = "app"
	defaultBuilCMD    = "go build -o $1"
)

// Config is the monorepo Service configuration
type Config struct {
	ServicePath string   `json:"service_path,omitempty"`
	RepoPath    string   `json:"repo_path,omitempty"`
	Extra       []string `json:"extra,omitempty"`
	BuildCMD    string   `json:"build_cmd,omitempty"`
	BinaryName  string   `json:"binary_name,omitempty"`
}

// AbsolutePath returns the absolute path of the services in the monorepo
func (c Config) AbsolutePath() string {
	return c.RepoPath + "/" + c.ServicePath
}

// ServiceDirs returns the directories of all services in the monorepo
func (c Config) ServiceDirs() ([]string, error) {
	cmdDirs, err := filepath.Glob(c.AbsolutePath())
	if err != nil {
		return nil, err
	}

	return cmdDirs, nil
}

// BuildArgs returns the command and arguments required to build a service
func (c Config) BuildArgs() (string, []string) {
	if c.BuildCMD != "" && c.BinaryName != "" {
		c.BuildCMD = defaultBuilCMD
		c.BinaryName = defaultBinaryName
	}

	arg := strings.Replace(c.BuildCMD, "$1", c.BinaryName, 1)
	args := strings.Split(arg, " ")
	if len(args) == 0 {
		return "", []string{}
	}

	return args[0], args[1:]
}

// Validate validates the config
func (c Config) Validate() error {
	if c.ServicePath == "" {
		return errors.New("service path is required")
	}

	return nil
}
