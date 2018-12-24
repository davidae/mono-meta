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
	defaultBuildName = "app"
	defaultBuilCMD   = "go build -o " + defaultBuildName

	UNDEFINED  = comment("UNDEFINED")
	MODIFIED   = comment("MODIFIED")
	UNMODIFIED = comment("UNMODIFIED")
	REMOVED    = comment("REMOVED")
	NEW        = comment("NEW")
)

// Config is the monorepo service configuration
type ServiceConfig struct {
	Path  string   `json:"path"`
	Extra []string `json:"exclude"`
	Cmd   string   `json:"cmd"`
}

type Service struct {
	Name      string `json:"name,omitempty"`
	Path      string `json:"path,omitempty"`
	Hash      string `json:"hash,omitempty"`
	Reference string `json:"reference,omitempty"`
}

type ServiceDiff struct {
	Name    string   `json:"name,omitempty"`
	Changed bool     `json:"changed,omitempty"`
	Comment comment  `json:"comment,omitempty"`
	Base    *Service `json:"base,omitempty"`
	Compare *Service `json:"compare,omitempty"`
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
			if d.Base.Hash != d.Compare.Hash {
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
		filename, err := buildPackage(d)
		if err != nil {
			return []*Service{}, nil
		}

		h, err := hashBuild(filename)
		if err != nil {
			return []*Service{}, err
		}

		services = append(services, &Service{
			Name:      getServiceName(absPath, filename),
			Hash:      h,
			Path:      filename,
			Reference: ref.Name().String(),
		})
	}

	return services, nil
}

func getServiceName(absPath, filePath string) string {
	abs := strings.Split(absPath, "/")
	file := strings.Split(filePath, "/")

	for i := range file {
		if file[i] != abs[i] {
			return file[i]
		}
	}

	return ""
}

func buildPackage(dir string) (string, error) {
	cmdName, cmdArgs := buildArgs(defaultBuilCMD)
	if cmdName == "" {
		return "", fmt.Errorf("invalid build args: '%s'", defaultBuilCMD)
	}

	buildCmd := exec.Command(cmdName, cmdArgs...)
	buildCmd.Dir = dir
	out, err := buildCmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrapf(err, "msg: %s", string(out))
	}

	return dir + "/" + defaultBuildName, nil
}

func buildArgs(s string) (string, []string) {
	args := strings.Split(s, " ")
	if len(args) == 0 {
		return "", []string{}
	}

	return args[0], args[1:]
}

func hashBuild(filename string) (string, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", md5.Sum(file)), nil
}
