# mono-meta
A CLI and API to retrieve service information from a Go based monorepo with a microservice architecture. 

[![GoDoc](https://godoc.org/github.com/davidae/mono-meta/mono?status.svg)](https://godoc.org/github.com/davidae/mono-meta/mono)
[![Build Status](https://travis-ci.org/davidae/mono-meta.svg "Travis CI status")](https://travis-ci.org/davidae/mono-meta)


The main purpose is to determine which microservices were modified, removed or added when comparing two different 
references, a reference can be a branch, release, tag or similar.
This can simplify the rollout on different stages - you can do _selective_ builds, tests and releases on a set of 
services and avoid doing an _all-or-nothing_ approach on your CI that can take a considerable amount of time. 

`mono-meta` will build all services locally, after cloning the repo or using an existing local repo, and do a checksum over the binaries. 
The checksum will determine if there was any changes from one branch to another, `mono-meta` therefore assumes that go builds are _deterministic_. A talk by [davecheney](https://go-talks.appspot.com/github.com/davecheney/presentations/reproducible-builds.slide#1) and a post by [filippo.io](https://blog.filippo.io/reproducing-go-binaries-byte-by-byte/) are two very good resources to continue reading about reproducible go builds.

## Install
```
$ go get -u github.com/davidae/mono-meta
```
Or get the latest binary release [here](https://github.com/davidae/mono-meta/releases)

## Dependencies and other requirements
* `go1.7` or higher. Go is needed to build binaries for each service on-the-fly to do checksums over the binaries to determine changes.
* `git`

A **structured** monorepo is required, it is important to know the path to the services directory and that they can all be listed, e.g.
```
services/*/cmd
code/go/services/
*
```
are all valid `services` path, but only one can exist and must be specified with a flag or in a config file. 
And each final/sub directory has to be able to build, e.g. `main.go` must exist. Check out this [sample repo](https://github.com/davidae/service-struct-repo) for an example.


## CLI
```
mono-meta [command]

Available Commands:
  diff        Get a diff summary of all services in a monorepo between two references
  help        Help about any command
  services    Get a service summary of all services in a monorepo

Flags:
  -h, --help   help for mono-meta
```
The commands are outputted to stdout in a JSON format, any errors goes to stderr.

### Usage
For a clean CLI command it is suggested to use a file (`--file`/`-f`) option, please see `sample.json` for references.
```
mono-meta diff -f sample.json -c add-new-service-and-change
mono-meta services -u git@github.com:davidae/service-struct-repo.git -s 'services/*/cmd'
```

### Flags
These are the common flags for both `diff` and `services` command.
```
  -e, --build-cmd string   build command when building services, a '$1' variable of the outputted binary is required (default "go build -o $1")
  -f, --file string        load configuration file - it will override any existing values set by a flag with same "key"
  -h, --help               help for diff
  -l, --local string       local URL of the git repository, use for specifying existing repo or where remote repo is to be cloned (default "/tmp/monorepo")
  -s, --services string    path pattern of where the services resides in the git repository
  -u, --url string         remote URL of the git repository, unnecessary to use if the repo is already locally
```

## API
Below is a sample how to use the API in a Go application
```go
package main

import (
  "fmt"

  "github.com/davidae/mono-meta/mono"
  "github.com/davidae/mono-meta/repo"
)

func main() {
  // See repo.NewLocal for using an already existing local git repo
  repo, err := repo.NewRemote("git@github.com:davidae/service-struct-repo.git")
  if err != nil {
    panic(err)
  }

  meta := mono.NewMonoMeta(repo, mono.Config{
    BuildCMD:    "go build -o $1",
    ServicePath: "services/*/cmd",
  })

  diffs, err := meta.Diff("master", "another-branch")
  if err != nil {
    panic(err)
  }

  fmt.Println(diffs)

  services, err := meta.Services("master")
  if err != nil {
    panic(err)
  }

  fmt.Println(services)

  // removes repo locally
  if err = repo.Close(); err != nil {
    panic(err)
  }
}
```
