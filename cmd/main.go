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
	pretty  bool

	directory = "/tmp/hello"
)

func main() {
	flag.StringVar(&config, "config", "", "descrip here")
	flag.StringVar(&url, "url", "", "descrip here")
	flag.StringVar(&base, "base", "master", "descrip here")
	flag.StringVar(&compare, "compare", "", "descrip here")
	flag.BoolVar(&pretty, "pretty", false, "descrip here")

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

	out, err := toJSON(pretty, diffs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		return
	}

	fmt.Fprint(os.Stdout, out)
}

func toJSON(pretty bool, diffs []service.ServiceDiff) (string, error) {
	if pretty {
		data, err := json.MarshalIndent(diffs, " ", "  ")
		if err != nil {
			return "", err
		}

		return string(data), nil
	}

	data, err := json.Marshal(diffs)
	if err != nil {
		return "", err
	}

	return string(data), nil
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
