package mono

import (
	"errors"
	"path/filepath"
	"strings"
)

const (
	DefaultBinaryName = "app"
	DefaultBuilCMD    = "go build -o $1"
)

// Config is the monorepo Service configuration
type Config struct {
	ServicePath string   `json:"service_path,omitempty"`
	RepoPath    string   `json:"repo_path,omitempty"`
	Extra       []string `json:"extra,omitempty"`
	BuildCMD    string   `json:"build_cmd,omitempty"`
	BinaryName  string   `json:"binary_name,omitempty"`
}

func (c Config) AbsolutePath() string {
	return c.RepoPath + "/" + c.ServicePath
}

func (c Config) ServiceDirs() ([]string, error) {
	cmdDirs, err := filepath.Glob(c.AbsolutePath())
	if err != nil {
		return nil, err
	}

	return cmdDirs, nil
}

func (c Config) BuildArgs() (string, []string) {
	arg := strings.Replace(c.BuildCMD, "$1", c.BinaryName, 1)
	args := strings.Split(arg, " ")
	if len(args) == 0 {
		return "", []string{}
	}

	return args[0], args[1:]
}

func (c Config) ServiceName(filepath string) string {
	abs := strings.Split(c.AbsolutePath(), "/")
	file := strings.Split(filepath, "/")

	for i := range file {
		if file[i] != abs[i] {
			return file[i]
		}
	}

	return ""
}

func (c Config) Validate() error {
	if c.ServicePath == "" {
		return errors.New("service path is required")
	}

	if c.BuildCMD != "" && c.BinaryName != "" {
		return nil
	}

	c.BuildCMD = DefaultBuilCMD
	c.BinaryName = DefaultBinaryName

	return nil
}
