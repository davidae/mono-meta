package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/davidae/mono-builder/git"
	"github.com/davidae/mono-builder/logger"
)

var (
	config  string
	repo    string
	release string
	origin  string
	target  string
	output  string
	debug   bool
)

func main() {
	flag.StringVar(&config, "config", "", "descrip here")
	flag.StringVar(&repo, "repo", "", "descrip here")
	flag.StringVar(&release, "release", "", "descrip here")
	flag.StringVar(&origin, "origin", "", "descrip here")
	flag.StringVar(&target, "target", "master", "descrip here")
	flag.StringVar(&output, "output", "stdout", "descrip here")
	flag.BoolVar(&debug, "debug", false, "desc here")

	flag.Parse()

	logger.Debug(debug)

	if !git.IsAvailable() {
		logger.Error(nil, "git was not found")
	}

	d, err := ioutil.ReadFile(config)
	if err != nil {
		logger.Error(err, fmt.Sprintf("unable to read file at '%s'", config))
		return
	}

	var cfg MonoConfig
	if err := json.Unmarshal(d, &cfg); err != nil {
		logger.Error(err, fmt.Sprintf("unable to unmarshal json file '%s", config))
		return
	}

	logger.Log("cloning " + repo)
	if err := git.Clone(repo); err != nil {
		logger.Error(err, fmt.Sprintf("failed to clone '%s'", repo))
		return
	}

	defer func() {
		logger.Log("cleaning up")
		if err := git.Cleanup(); err != nil {
			logger.Error(err, fmt.Sprintf("failed to clean up temp folder"))
			return
		}
	}()

	diff, err := git.Diff(target, origin)
	if err != nil {
		logger.Error(err, "diff failed")
		return
	}

	fmt.Println(diff)
}

// MonoConfig is the monorepo service configuration
type MonoConfig struct {
	Path    string   `json:"path,omitempty"`
	Exclude []string `json:"exclude,omitempty"`
	Extra   []string `json:"extra,omitempty"`
}
