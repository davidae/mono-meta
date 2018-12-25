package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/davidae/mono-builder/mono"
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

	m, err := mono.NewMonoMeta(url, directory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		return
	}

	defer func() {
		if err := m.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			return
		}
	}()

	diffs, err := m.Diff(cfg, base, compare)
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

func toJSON(pretty bool, diffs []mono.ServiceDiff) (string, error) {
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

func parseConfig(config string) (mono.Config, error) {
	d, err := ioutil.ReadFile(config)
	if err != nil {
		return mono.Config{}, err

	}

	var cfg mono.Config
	if err := json.Unmarshal(d, &cfg); err != nil {
		return mono.Config{}, err
	}

	if cfg.BuildCMD == "" || cfg.BinaryName == "" {
		cfg.BuildCMD = mono.DefaultBuilCMD
		cfg.BinaryName = mono.DefaultBinaryName
	}

	return cfg, nil
}
