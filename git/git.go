package git

import (
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// Directory where the repo is cloned into
const Directory = "/tmp/repo"

// IsAvailable checks is git is available on the executing environement
func IsAvailable() bool {
	return exec.Command("git", "version").Run() == nil
}

// Clone clones a repo into a temporary directory
func Clone(repo string) error {
	msg, err := exec.Command(
		"git",
		"clone",
		repo,
		Directory,
	).CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "msg: %s", string(msg))
	}

	return nil
}

// Cleanup removes any repos cloned
func Cleanup() error {
	return os.RemoveAll(Directory)
}

// Diff returns the git diff between a target and origin branch
func Diff(target, origin string) ([]string, error) {
	out, err := exec.Command(
		"git",
		"--git-dir",
		Directory+"/.git",
		"diff",
		"--name-only",
		target+"..."+origin,
	).CombinedOutput()

	if err != nil {
		return []string{}, errors.Wrapf(err, "msg: %s", string(out))
	}

	var res []string
	for _, f := range strings.Split(string(out), "\n") {
		f = strings.Trim(f, " ")
		if f != "" {
			res = append(res, f)
		}
	}

	return res, nil
}

func Checkout(branch string) error {
	out, err := exec.Command(
		"git",
		"--git-dir",
		Directory+"/.git",
		"checkout",
	)
	if err =! nil {
		return errors.Wrapf(err, "msg: %s", string(out))
	}

	return nil
}
