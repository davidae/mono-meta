package mono

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/davidae/mono-meta/repo"

	"github.com/pkg/errors"
)

// Comment is a comment about the status a of service
type comment string

const (
	// MODIFIED is an enum describing that a service has been modified
	MODIFIED = comment("modified")
	// UNMODIFIED is an enum describing that a service is unmodified
	UNMODIFIED = comment("unmodified")
	// REMOVED is an enum describing that a service has been removed
	REMOVED = comment("removed")
	// NEW is an enum describing that a service is new
	NEW = comment("new")
)

// Service is a description of a service
type Service struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Checksum  string `json:"checksum"`
	Reference string `json:"reference"`
}

// ServiceDiff is a description of a service compared between two references
type ServiceDiff struct {
	Name    string   `json:"name"`
	Changed bool     `json:"changed"`
	Comment comment  `json:"comment"`
	Base    *Service `json:"base"`
	Compare *Service `json:"compare"`
}

// Meta represents the metadata of a monorepo
type Meta struct {
	repo   repo.Repository
	config Config
}

// NewMonoMeta returns a new Meta instance
func NewMonoMeta(repo repo.Repository, c Config) *Meta {
	return &Meta{repo: repo, config: c}
}

// Diff returns a slice of services compared across two git references
func (m Meta) Diff(base, compare string) ([]*ServiceDiff, error) {
	bas, err := m.Services(base)
	if err != nil {
		return []*ServiceDiff{}, errors.Wrapf(err, "Diff error, failed to get services from %s", base)
	}

	com, err := m.Services(compare)
	if err != nil {
		return []*ServiceDiff{}, errors.Wrapf(err, "Diff error, failed to get services from %s", compare)
	}

	services := map[string]*ServiceDiff{}
	for _, c := range com {
		services[c.Name] = &ServiceDiff{
			Name:    c.Name,
			Compare: c,
		}
	}

	for _, b := range bas {
		d, ok := services[b.Name]
		if !ok {
			services[b.Name] = &ServiceDiff{Base: b}
			continue
		}

		d.Base = b
	}

	diffs := []*ServiceDiff{}
	for _, s := range services {
		if s.Base == nil && s.Compare != nil {
			s.Changed = true
			s.Comment = NEW
		}
		if s.Base != nil && s.Compare == nil {
			s.Changed = true
			s.Comment = REMOVED
			s.Name = s.Base.Name
		}
		if s.Base != nil && s.Compare != nil {
			if s.Base.Checksum != s.Compare.Checksum {
				s.Changed = true
				s.Comment = MODIFIED
			} else {
				s.Changed = false
				s.Comment = UNMODIFIED
			}
		}

		diffs = append(diffs, s)
	}

	sort.Slice(diffs, func(i, j int) bool {
		return diffs[i].Name < diffs[j].Name
	})

	return diffs, nil
}

// Services returns all services for a given reference in the monorepo
func (m Meta) Services(reference string) ([]*Service, error) {
	ref, err := m.repo.Checkout(reference)
	if err != nil {
		return nil, err
	}

	cmdDirs, err := m.serviceDirs()
	if err != nil {
		return nil, errors.Wrap(err, "could not find service directories")
	}

	services := make([]*Service, 0, len(cmdDirs))
	for _, d := range cmdDirs {
		filename, err := m.buildPackage(d)
		if err != nil {
			return []*Service{}, errors.Wrap(err, "could not build package")
		}

		csum, err := checksum(filename)
		if err != nil {
			return []*Service{}, errors.Wrap(err, "could not checksum binary")
		}

		services = append(services, &Service{
			Name:      m.serviceName(filename),
			Checksum:  csum,
			Path:      filename,
			Reference: ref,
		})
	}

	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})

	return services, nil
}

func (m Meta) serviceName(filepath string) string {
	abs := strings.Split(m.servicesPath(), "/")
	file := strings.Split(filepath, "/")

	for i := range file {
		if file[i] != abs[i] {
			return file[i]
		}
	}

	return ""
}

func (m Meta) buildPackage(dir string) (string, error) {
	c, args := m.config.BuildArgs()
	cmd := exec.Command(c, args...)
	cmd.Dir = dir

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrapf(err, "%s", string(out))
	}

	return dir + "/" + binary, nil
}

func checksum(filename string) (string, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", md5.Sum(file)), nil
}

func (m Meta) servicesPath() string {
	return m.repo.LocalPath() + "/" + m.config.ServicePath
}

func (m Meta) serviceDirs() ([]string, error) {
	cmdDirs, err := filepath.Glob(m.servicesPath())
	if err != nil {
		return nil, err
	}

	return cmdDirs, nil
}
