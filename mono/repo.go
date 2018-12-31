package mono

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

const tempDir = "/tmp/monorepo"

// Repository is an interface
type Repository interface {
	Checkout(ref string) (string, error)
	LocalPath() string
	Close() error
}

type remote struct {
	repo *git.Repository
	path string
}

type local struct {
	repo *git.Repository
	path string
}

// NewRemote clones a remote git repository
func NewRemote(url string) (Repository, error) {
	r, err := git.PlainClone(tempDir, false, &git.CloneOptions{
		URL: url,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to clone repo into '%s'", tempDir)
	}

	return &remote{repo: r, path: tempDir}, nil
}

func (r *remote) LocalPath() string {
	return r.path
}

func (r *remote) Checkout(ref string) (string, error) {
	return checkout(r.repo, ref)
}

func (r *remote) Close() error {
	return os.RemoveAll(r.path)
}

// NewLocal uses a local git repository
func NewLocal(path string) (Repository, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open local git repo (%s)", path)

	}

	return &local{repo: r, path: path}, nil
}

func (l *local) LocalPath() string {
	return l.path
}

func (l *local) Checkout(ref string) (string, error) {
	return checkout(l.repo, ref)
}

func (l *local) Close() error {
	return nil
}

func checkout(repo *git.Repository, ref string) (string, error) {
	w, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	refs, err := repo.References()
	if err != nil {
		return "", err
	}

	var reference *plumbing.Reference

	refs.ForEach(func(r *plumbing.Reference) error {
		if strings.HasSuffix(r.Name().String(), ref) {
			reference = r
		}
		return nil
	})

	if reference == nil {
		return "", fmt.Errorf("could not find ref/branch: %s", ref)
	}

	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.Hash(reference.Hash()),
	})
	if err != nil {
		return "", err
	}

	return reference.Name().String(), nil
}
