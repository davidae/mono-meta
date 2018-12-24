package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/davidae/mono-builder/git"
	"github.com/davidae/mono-builder/logger"
	"github.com/davidae/mono-builder/service"
)

var (
	config  string
	url     string
	release string
	origin  string
	target  string
	output  string
	debug   bool

	directory = "/tmp/hello"
)

func main() {
	flag.StringVar(&config, "config", "", "descrip here")
	flag.StringVar(&url, "url", "", "descrip here")
	flag.StringVar(&release, "release", "", "descrip here")
	flag.StringVar(&origin, "origin", "", "descrip here")
	flag.StringVar(&target, "target", "master", "descrip here")
	flag.StringVar(&output, "output", "stdout", "descrip here")
	flag.BoolVar(&debug, "debug", false, "desc here")

	flag.Parse()

	logger.Debug(debug)

	d, err := ioutil.ReadFile(config)
	if err != nil {
		logger.Error(err, fmt.Sprintf("unable to read file at '%s'", config))
		return
	}

	var cfg service.ServiceConfig
	if err := json.Unmarshal(d, &cfg); err != nil {
		logger.Error(err, fmt.Sprintf("unable to unmarshal json file '%s", config))
		return
	}

	logger.Log("cloning " + url + " into " + directory)
	r, err := git.Clone(url, directory)
	if err != nil {
		fmt.Printf("err!!! %s\n", err)
		return
	}

	defer func() {
		logger.Log("cleaning up")
		if err := r.Cleanup(); err != nil {
			logger.Error(err, fmt.Sprintf("failed to clean up temp folder"))
			return
		}
	}()

	diffs, err := service.Diff(r, cfg, origin, target)
	if err != nil {
		fmt.Printf("err!!! %s\n", err)
		return
	}

	fmt.Printf("diffs: %#v\n", diffs)

}
