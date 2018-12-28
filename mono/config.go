package mono

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// Binary is the default name out an outputted binary
	binary = "app"
)

// Config is the monorepo service configuration
type Config struct {
	ServicePath string `json:"services,omitempty"`
	BuildCMD    string `json:"cmd,omitempty"`
}

// BuildArgs returns the command and arguments required to build a service
func (c Config) BuildArgs() (string, []string) {
	arg := strings.Replace(c.BuildCMD, "$1", binary, 1)
	args := strings.Split(arg, " ")
	if len(args) == 0 {
		return "", []string{}
	}

	return args[0], args[1:]
}

// Validate validates the config
func (c Config) Validate() error {
	if c.ServicePath == "" {
		return errors.New("services path is required")
	}

	if !strings.Contains(c.BuildCMD, "-o $1") {
		return fmt.Errorf("build command (%s) must output to arg $1, e.g. '-o $1'", c.BuildCMD)
	}

	return nil
}
