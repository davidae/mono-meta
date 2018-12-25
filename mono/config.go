package mono

import (
	"path/filepath"
	"strings"
)

const (
	DefaultBinaryName = "app"
	DefaultBuilCMD    = "go build -o $1"
)

// Cfg is the monorepo Service configuration
type Cfg struct {
	ServicePath string   `json:"service_path,omitempty"`
	RepoPath    string   `json:"repo_path,omitempty"`
	Extra       []string `json:"extra,omitempty"`
	BuildCMD    string   `json:"build_cmd,omitempty"`
	BinaryName  string   `json:"binary_name,omitempty"`
}

func (c Cfg) AbsolutePath() string {
	return c.RepoPath + "/" + c.ServicePath
}

func (c Cfg) ServiceDirs() ([]string, error) {
	cmdDirs, err := filepath.Glob(c.AbsolutePath())
	if err != nil {
		return nil, err
	}

	return cmdDirs, nil
}

func (c Cfg) BuildArgs() (string, []string) {
	arg := strings.Replace(c.BuildCMD, "$1", c.BinaryName, 1)
	args := strings.Split(arg, " ")
	if len(args) == 0 {
		return "", []string{}
	}

	return args[0], args[1:]
}

func (c Cfg) ServiceName(filepath string) string {
	abs := strings.Split(c.AbsolutePath(), "/")
	file := strings.Split(filepath, "/")

	for i := range file {
		if file[i] != abs[i] {
			return file[i]
		}
	}

	return ""
}

func (c Cfg) Validate() error {
	if c.BuildCMD != "" && c.BinaryName != "" {
		return nil
	}

	c.BuildCMD = DefaultBuilCMD
	c.BinaryName = DefaultBinaryName

	return nil
}
