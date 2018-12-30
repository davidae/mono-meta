package mono_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/davidae/mono-meta/mono"
	"github.com/davidae/mono-meta/repo"
)

const (
	repo1 = "../test/mock-repo-1"
	repo2 = "../test/mock-repo-2"
)

var repoPath string

type MockRepo struct {
	switchRepo func() string
}

func (m MockRepo) Checkout(ref string) (string, error) {
	repoPath = m.switchRepo()
	return "test/ref/master", nil
}

func (m MockRepo) LocalPath() string {
	return repoPath
}

func (m MockRepo) Close() error {
	return nil
}

func TestMono(t *testing.T) {
	tests := []struct {
		scenario string
		repo     repo.Repository
		fn       func(*testing.T, repo.Repository)
	}{
		{
			scenario: "test services - success",
			repo:     MockRepo{switchRepo: returnString(repo1)},
			fn:       testServices,
		},
		{
			scenario: "test dff - success",
			repo:     MockRepo{switchRepo: alternateRepos()},
			fn:       testDiff,
		},
	}

	for _, tc := range tests {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.fn(t, tc.repo)
			if err := cleanUp(); err != nil {
				t.Fatalf("failed to clean up test, mock binaries have been produced: %s", err)
			}
		})
	}
}

func testServices(t *testing.T, r repo.Repository) {
	m := mono.NewMonoMeta(
		r,
		mono.Config{
			BuildCMD:    "go build -o $1",
			ServicePath: "services/*",
		})

	services, err := m.Services("irrelevant")
	assert.NoError(t, err)
	assert.Equal(t, len(services), 2)
	assert.Equal(t, "service-1", services[0].Name)
	assert.Equal(t, "../test/mock-repo-1/services/service-1/app", services[0].Path)
	assert.Equal(t, "bc9ee478b6f14bea43bd1524a3031018", services[0].Checksum)
	assert.Equal(t, "test/ref/master", services[0].Reference)

	assert.Equal(t, "service-2", services[1].Name)
	assert.Equal(t, "../test/mock-repo-1/services/service-2/app", services[1].Path)
	assert.Equal(t, "c585bff21820859a197582e2b4e1f1ce", services[1].Checksum)
	assert.Equal(t, "test/ref/master", services[1].Reference)
}

func testDiff(t *testing.T, r repo.Repository) {
	m := mono.NewMonoMeta(
		r,
		mono.Config{
			BuildCMD:    "go build -o $1",
			ServicePath: "services/*",
		})

	services, err := m.Diff("fi", "ne")
	assert.NoError(t, err)
	assert.Equal(t, len(services), 3)

	assert.Equal(t, "service-1", services[0].Name)
	assert.True(t, services[0].Changed)
	assert.Equal(t, "removed", string(services[0].Comment))
	assert.Empty(t, services[0].Compare)
	assert.NotEmpty(t, services[0].Base)

	assert.Equal(t, "service-2", services[1].Name)
	assert.True(t, services[1].Changed)
	assert.Equal(t, "modified", string(services[1].Comment))
	assert.NotEmpty(t, services[1].Compare)
	assert.NotEmpty(t, services[1].Base)

	assert.Equal(t, "service-3", services[2].Name)
	assert.True(t, services[2].Changed)
	assert.Equal(t, "new", string(services[2].Comment))
	assert.NotEmpty(t, services[2].Compare)
	assert.Empty(t, services[2].Base)
}

func returnString(s string) func() string {
	return func() string { return s }
}

func alternateRepos() func() string {
	flip := true

	return func() string {
		defer func() { flip = !flip }()
		if flip {
			return repo1
		}
		return repo2
	}
}

func cleanUp() error {
	for _, r := range []string{repo1, repo2} {
		cmdDirs, err := filepath.Glob(r + "/services/*")
		if err != nil {
			return err
		}

		for _, d := range cmdDirs {
			_, err := os.Stat(d + "/" + mono.Binary)
			if os.IsNotExist(err) {
				continue
			}

			if err := os.Remove(d + "/" + mono.Binary); err != nil {
				return err
			}
		}
	}

	return nil
}
