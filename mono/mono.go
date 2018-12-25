package mono

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
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

// Config is the monorepo Service configuration
type Config struct {
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

type MonoMeta struct {
	repo      *git.Repository
	directory string
}

func NewMonoMeta(repoURL, tempDir string) (MonoMeta, error) {
	r, err := git.PlainClone(tempDir, false, &git.CloneOptions{
		URL: repoURL,
	})
	if err != nil {
		return MonoMeta{}, err
	}

	return MonoMeta{repo: r, directory: tempDir}, nil
}

func (m MonoMeta) Diff(cfg Config, base, compare string) ([]ServiceDiff, error) {
	bas, err := m.GetServices(cfg, base)
	if err != nil {
		return []ServiceDiff{}, errors.Wrapf(err, "Diff error, failed to get services from %s", base)
	}

	com, err := m.GetServices(cfg, compare)
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
func (m MonoMeta) GetServices(cfg Config, reference string) ([]*Service, error) {
	ref, err := m.Checkout(reference)
	if err != nil {
		return nil, err
	}

	absPath := m.directory + "/" + cfg.Path
	cmdDirs, err := filepath.Glob(absPath)
	if err != nil {
		return nil, err
	}

	services := make([]*Service, 0, len(cmdDirs))
	for _, d := range cmdDirs {
		filename, err := buildPackage(cfg, d)
		if err != nil {
			return []*Service{}, err
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

func buildPackage(cfg Config, dir string) (string, error) {
	cmdName, cmdArgs := buildArgs(cfg)
	if cmdName == "" {
		return "", fmt.Errorf("invalid build args: '%s'", cfg.BuildCMD)
	}

	buildCmd := exec.Command(cmdName, cmdArgs...)
	buildCmd.Dir = dir
	out, err := buildCmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrapf(err, "%s", string(out))
	}

	return dir + "/" + DefaultBinaryName, nil
}

func buildArgs(cfg Config) (string, []string) {
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
