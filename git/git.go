package git

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

type Repo struct {
	Repo      *git.Repository
	Directory string
}

func Clone(url, directory string) (Repo, error) {
	r, err := git.PlainClone(directory, false, &git.CloneOptions{
		URL: url,
	})
	if err != nil {
		return Repo{}, err
	}

	return Repo{Repo: r, Directory: directory}, nil
}

func (r *Repo) Checkout(b string) (*plumbing.Reference, error) {
	w, err := r.Repo.Worktree()
	if err != nil {
		return nil, err
	}

	refs, err := r.Repo.References()
	if err != nil {
		return nil, err
	}

	var reference *plumbing.Reference

	refs.ForEach(func(r *plumbing.Reference) error {
		if strings.HasSuffix(r.Name().String(), b) {
			reference = r
		}
		return nil
	})

	if reference == nil {
		return nil, fmt.Errorf("could not find ref/branch: %s", b)
	}

	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.Hash(reference.Hash()),
	})
	if err != nil {
		return nil, err
	}

	return reference, nil
}

func (r *Repo) Cleanup() error {
	return os.RemoveAll(r.Directory)
}
