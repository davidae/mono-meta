package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/davidae/mono-meta/cmd"
	"github.com/davidae/mono-meta/mono"
)

// Flags
var (
	file        string
	local       string
	url         string
	buildCMD    string
	clonePath   string
	servicePath string
	compare     string
	base        string
	ref         string
)

var (
	monoConfig mono.Config
	repoConfig RepoConfig

	root     *cobra.Command
	diff     *cobra.Command
	services *cobra.Command
)

// RepoConfig holds all configuration required for retrieve a repo remotely and locally
type RepoConfig struct {
	Path string `json:"local,omitempty"`
	URL  string `json:"url,omitempty"`
}

func init() {
	root = &cobra.Command{Use: "mono-meta", TraverseChildren: true, Args: cobra.ExactValidArgs(1)}
	assignFlags := func(f *pflag.FlagSet) {
		f.StringVarP(&file, "file", "f", "", "load configuration file - it will override any existing values set by a flag with same \"key\"")
		f.StringVarP(&local, "local", "l", "", "local path to a git repository")
		f.StringVarP(&url, "url", "u", "", "remote URL to a git repository")
		f.StringVarP(&buildCMD, "build-cmd", "e", "go build -o $1", "go build command for all services, a '$1' variable of the outputted binary is required")
		f.StringVarP(&servicePath, "services", "s", "", "path (pattern) where the microservices resides in the monorepo")
	}

	diff = cmd.Diff()
	assignFlags(diff.Flags())
	diff.Flags().StringVarP(&base, "base", "b", "master", "branch to be used, as the base, in the comparison of the diff in the monorepo")
	diff.Flags().StringVarP(&compare, "compare", "c", "", "branch to be used to compare with the base in the monorepo")
	diff.MarkFlagRequired("compare")
	diff.MarkFlagRequired("services")

	services = cmd.Services()
	assignFlags(services.Flags())
	services.Flags().StringVarP(&ref, "branch", "b", "master", "branch to be used for the summary of all services in the monorepo")
	services.MarkFlagRequired("services")
}

func main() {
	diff.Run = func(cmd *cobra.Command, args []string) {
		err := loadConfig(file, &monoConfig, &repoConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			return
		}

		d, err := getDiff(monoConfig, repoConfig, base, compare)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			return
		}

		out, err := toJSON(true, d)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			return
		}

		fmt.Fprint(os.Stdout, out)
	}

	services.Run = func(cmd *cobra.Command, args []string) {
		err := loadConfig(file, &monoConfig, &repoConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			return
		}
		s, err := getServices(monoConfig, repoConfig, ref)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			return
		}

		out, err := toJSON(true, s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			return
		}

		fmt.Fprint(os.Stdout, out)
	}

	root.AddCommand(services, diff)
	root.Execute()
}

func toJSON(pretty bool, resp interface{}) (string, error) {
	if pretty {
		data, err := json.MarshalIndent(resp, " ", "  ")
		if err != nil {
			return "", err
		}

		return string(data), nil
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func loadConfig(config string, m *mono.Config, r *RepoConfig) error {
	monoConfig.ServicePath = servicePath
	monoConfig.BuildCMD = buildCMD

	repoConfig.URL = url
	repoConfig.Path = local

	if config != "" {
		d, err := ioutil.ReadFile(config)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(d, m); err != nil {
			return err
		}

		if err := json.Unmarshal(d, r); err != nil {
			return err
		}
	}

	if repoConfig.URL == "" && repoConfig.Path == "" {
		return errors.New("--url or --local flag must be set")
	}

	return monoConfig.Validate()
}

func getDiff(mCfg mono.Config, rCfg RepoConfig, base, compare string) ([]*mono.ServiceDiff, error) {
	m, r, err := getMeta(mCfg, rCfg)
	if err != nil {
		return []*mono.ServiceDiff{}, err
	}

	defer r.Close()

	diffs, err := m.Diff(base, compare)
	if err != nil {
		return []*mono.ServiceDiff{}, err
	}

	return diffs, nil
}

func getServices(mCfg mono.Config, rCfg RepoConfig, ref string) ([]*mono.Service, error) {
	m, r, err := getMeta(mCfg, rCfg)
	if err != nil {
		return []*mono.Service{}, err
	}

	defer r.Close()

	services, err := m.Services(ref)
	if err != nil {
		return []*mono.Service{}, err
	}

	return services, nil
}

func getMeta(mCfg mono.Config, rCfg RepoConfig) (*mono.Meta, mono.Repository, error) {
	var (
		r   mono.Repository
		err error
	)

	if rCfg.URL == "" {
		r, err = mono.NewLocal(rCfg.Path)
		if err != nil {
			return nil, nil, err
		}
	} else {
		r, err = mono.NewRemote(rCfg.URL)
		if err != nil {
			return nil, nil, err
		}
	}

	return mono.NewMeta(r, mCfg), r, nil
}
