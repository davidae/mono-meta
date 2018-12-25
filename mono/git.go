package mono

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Checkout checks out a branch, tag or other reference in the mono-repo
func (m Meta) Checkout(b string) (*plumbing.Reference, error) {
	w, err := m.repo.Worktree()
	if err != nil {
		return nil, err
	}

	refs, err := m.repo.References()
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

// Close clears any temporary files or directories creates
func (m Meta) Close() error {
	return os.RemoveAll(m.config.RepoPath)
}
