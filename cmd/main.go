package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/davidae/mono-builder/git"
	"github.com/davidae/mono-builder/service"
)

var (
	config  string
	url     string
	base    string
	compare string
	output  string

	directory = "/tmp/hello"
)

func main() {
	flag.StringVar(&config, "config", "", "descrip here")
	flag.StringVar(&url, "url", "", "descrip here")
	flag.StringVar(&base, "base", "", "descrip here")
	flag.StringVar(&compare, "compare", "master", "descrip here")
	flag.StringVar(&output, "output", "stdout", "descrip here")

	flag.Parse()

	cfg, err := parseConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		return
	}

	r, err := git.Clone(url, directory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		return
	}

	defer func() {
		if err := r.Cleanup(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			return
		}
	}()

	diffs, err := service.Diff(r, cfg, base, compare)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		return
	}

	d, err := json.MarshalIndent(diffs, " ", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		return
	}

	fmt.Fprint(os.Stdout, string(d))
}

func parseConfig(config string) (service.ServiceConfig, error) {
	d, err := ioutil.ReadFile(config)
	if err != nil {
		return service.ServiceConfig{}, err

	}

	var cfg service.ServiceConfig
	if err := json.Unmarshal(d, &cfg); err != nil {
		return service.ServiceConfig{}, err
	}

	if cfg.BuildCMD == "" || cfg.BinaryName == "" {
		cfg.BuildCMD = service.DefaultBuilCMD
		cfg.BinaryName = service.DefaultBinaryName
	}

	return cfg, nil
}
