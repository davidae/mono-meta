package service

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/davidae/mono-builder/git"
	"github.com/pkg/errors"
)

type comment string

const (
	DefaultBinaryName = "app"
	DefaultBuilCMD    = "go build -o " + DefaultBinaryName

	UNDEFINED  = comment("UNDEFINED")
	MODIFIED   = comment("MODIFIED")
	UNMODIFIED = comment("UNMODIFIED")
	REMOVED    = comment("REMOVED")
	NEW        = comment("NEW")
)

// ServiceConfig is the monorepo service configuration
type ServiceConfig struct {
	Path       string   `json:"path"`
	Extra      []string `json:"extra"`
	BuildCMD   string   `json:"build_cmd"`
	BinaryName string   `json:"binary_name"`
}

type Service struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Checksum  string `json:"checksum"`
	Reference string `json:"reference"`
}

type ServiceDiff struct {
	Name    string   `json:"name"`
	Changed bool     `json:"changed"`
	Comment comment  `json:"comment"`
	Base    *Service `json:"base"`
	Compare *Service `json:"compare"`
}

func Diff(r git.Repo, cfg ServiceConfig, base, compare string) ([]ServiceDiff, error) {

	bas, err := Get(r, cfg, base)
	if err != nil {
		return []ServiceDiff{}, errors.Wrapf(err, "Diff error, failed to get services from base: %s", bas)
	}

	com, err := Get(r, cfg, compare)
	if err != nil {
		return []ServiceDiff{}, errors.Wrapf(err, "Diff error, failed to get services from compare: %s", com)
	}

	m := map[string]*ServiceDiff{}
	for _, c := range com {
		m[c.Name] = &ServiceDiff{
			Name:    c.Name,
			Compare: c,
		}
	}

	for _, b := range bas {
		d, ok := m[b.Name]
		if !ok {
			m[b.Name] = &ServiceDiff{Base: b}
			continue
		}

		d.Base = b

	}

	diffs := []ServiceDiff{}
	for _, d := range m {
		if d.Base == nil && d.Compare != nil {
			d.Changed = true
			d.Comment = NEW
		}
		if d.Base != nil && d.Compare == nil {
			d.Changed = true
			d.Comment = REMOVED
		}
		if d.Base != nil && d.Compare != nil {
			if d.Base.Checksum != d.Compare.Checksum {
				d.Changed = true
				d.Comment = MODIFIED
			} else {
				d.Changed = false
				d.Comment = UNMODIFIED
			}
		}

		diffs = append(diffs, *d)
	}

	return diffs, nil
}

func Get(r git.Repo, cfg ServiceConfig, reference string) ([]*Service, error) {
	ref, err := r.Checkout(reference)
	if err != nil {
		return nil, err
	}

	absPath := r.Directory + "/" + cfg.Path
	cmdDirs, err := filepath.Glob(absPath)
	if err != nil {
		return nil, err
	}

	services := make([]*Service, 0, len(cmdDirs))
	for _, d := range cmdDirs {
		filename, err := buildPackage(cfg, d)
		if err != nil {
			return []*Service{}, nil
		}

		csum, err := checksumBuild(filename)
		if err != nil {
			return []*Service{}, err
		}

		services = append(services, &Service{
			Name:      serviceName(absPath, filename),
			Checksum:  csum,
			Path:      filename,
			Reference: ref.Name().String(),
		})
	}

	return services, nil
}

func serviceName(absPath, filePath string) string {
	abs := strings.Split(absPath, "/")
	file := strings.Split(filePath, "/")

	for i := range file {
		if file[i] != abs[i] {
			return file[i]
		}
	}

	return ""
}

func buildPackage(cfg ServiceConfig, dir string) (string, error) {
	cmdName, cmdArgs := buildArgs(cfg)
	if cmdName == "" {
		return "", fmt.Errorf("invalid build args: '%s'", DefaultBuilCMD)
	}

	buildCmd := exec.Command(cmdName, cmdArgs...)
	buildCmd.Dir = dir
	out, err := buildCmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrapf(err, "msg: %s", string(out))
	}

	return dir + "/" + DefaultBinaryName, nil
}

func buildArgs(cfg ServiceConfig) (string, []string) {
	arg := strings.Replace(cfg.BuildCMD, "$1", cfg.BinaryName, 1)
	args := strings.Split(arg, " ")
	if len(args) == 0 {
		return "", []string{}
	}

	return args[0], args[1:]
}

func checksumBuild(filename string) (string, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", md5.Sum(file)), nil
}
