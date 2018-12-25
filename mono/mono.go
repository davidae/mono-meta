package mono

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
)

type comment string

const (
	UNDEFINED  = comment("UNDEFINED")
	MODIFIED   = comment("MODIFIED")
	UNMODIFIED = comment("UNMODIFIED")
	REMOVED    = comment("REMOVED")
	NEW        = comment("NEW")
)

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

type Meta struct {
	repo   *git.Repository
	config Cfg
}

func NewMonoMeta(repoURL string, c Cfg) (Meta, error) {
	r, err := git.PlainClone(c.RepoPath, false, &git.CloneOptions{
		URL: repoURL,
	})
	if err != nil {
		return Meta{}, err
	}

	if err = c.Validate(); err != nil {
		return Meta{}, err
	}

	return Meta{repo: r, config: c}, nil
}

func (m Meta) Diff(base, compare string) ([]ServiceDiff, error) {
	bas, err := m.GetServices(base)
	if err != nil {
		return []ServiceDiff{}, errors.Wrapf(err, "Diff error, failed to get services from %s", base)
	}

	com, err := m.GetServices(compare)
	if err != nil {
		return []ServiceDiff{}, errors.Wrapf(err, "Diff error, failed to get services from %s", compare)
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

	diffs := []ServiceDiff{}
	for _, s := range services {
		if s.Base == nil && s.Compare != nil {
			s.Changed = true
			s.Comment = NEW
		}
		if s.Base != nil && s.Compare == nil {
			s.Changed = true
			s.Comment = REMOVED
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

		diffs = append(diffs, *s)
	}

	return diffs, nil
}

// GetServices returns all services for a given reference
func (m Meta) GetServices(reference string) ([]*Service, error) {
	ref, err := m.Checkout(reference)
	if err != nil {
		return nil, err
	}

	cmdDirs, err := m.config.ServiceDirs()
	if err != nil {
		return nil, errors.Wrap(err, "could not find service directories")
	}

	services := make([]*Service, 0, len(cmdDirs))
	for _, d := range cmdDirs {
		filename, err := m.buildPackage(d)
		if err != nil {
			return []*Service{}, errors.Wrap(err, "could not build package")
		}

		csum, err := checksumBuild(filename)
		if err != nil {
			return []*Service{}, errors.Wrap(err, "could not checksum binary")
		}

		services = append(services, &Service{
			Name:      m.config.ServiceName(filename),
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

func (m Meta) buildPackage(dir string) (string, error) {
	cmdName, cmdArgs := m.config.BuildArgs()
	buildCmd := exec.Command(cmdName, cmdArgs...)
	buildCmd.Dir = dir

	out, err := buildCmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrapf(err, "%s", string(out))
	}

	return dir + "/" + DefaultBinaryName, nil
}

func checksumBuild(filename string) (string, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", md5.Sum(file)), nil
}
