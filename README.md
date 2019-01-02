# mono-meta
[![GoDoc](https://godoc.org/github.com/davidae/mono-meta/mono?status.svg)](https://godoc.org/github.com/davidae/mono-meta/mono)
[![Build Status](https://travis-ci.com/davidae/mono-meta.svg "Travis CI status")](https://travis-ci.com/davidae/mono-meta)

A CLI and API to retrieve service information from a Go based monorepo with a microservice architecture. 

The main purpose is to determine which microservices were modified, removed or added when comparing two different 
references, a reference can be a branch, release, tag or similar.
This can simplify the rollout on different stages - you can do _selective_ builds, tests and releases on a set of 
services and avoid doing an _all-or-nothing_ approach on your CI that can take a considerable amount of time. 

`mono-meta` assumes that go builds are deterministic and reproducible.
 A talk by [davecheney](https://go-talks.appspot.com/github.com/davecheney/presentations/reproducible-builds.slide#1) and 
a post by [filippo.io](https://blog.filippo.io/reproducing-go-binaries-byte-by-byte/) are two very good resources to continue 
reading about reproducible go builds.

## Install
```
$ go get -u github.com/davidae/mono-meta
```
Or get the latest binary release [here](https://github.com/davidae/mono-meta/releases)

## Dependencies and other requirements
* `go1.7` or higher.
* `git`

A **structured** monorepo is required, it is important to know the path to the services directory and that they can all be listed, e.g.
```
- services/*/cmd
- code/go/services/
- *
```
are all valid `services` path, but only one can exist and must be specified with a flag or in a config file. 
And each final/sub directory has to be able to build, e.g. `main.go` must exist. Check out this [sample repo](https://github.com/davidae/service-struct-repo) for an example.


## CLI
```
Usage:
  mono-meta [command]

Available Commands:
  diff        Get a diff summary of all services in a monorepo between two references
  help        Help about any command
  services    Get a summary of all services in a monorepo

Flags:
  -h, --help   help for mono-meta

Use "mono-meta [command] --help" for more information about a command.
```
The commands are outputted to stdout in a JSON format, any errors goes to stderr.

### Usage
For a clean CLI command it is suggested to use a file (`--file`/`-f`) option, please see `config.sample.json` for references.

### Flags
These are the common flags for both `diff` and `services` command.
```
  -e, --build-cmd string   go build command for all services, a '$1' variable of the outputted binary is required (default "go build -o $1")
  -f, --file string        load configuration file - it will override any existing values set by a flag with same "key"
  -h, --help               help for diff
  -l, --local string       local path to a git repository
  -u, --url string         remote URL to a git repository
  -s, --services string    path (pattern) where the microservices resides in the monorepo
```

### Example
```
$ mono-meta diff -u git@github.com:davidae/service-struct-repo.git -c add-new-service-and-change -s 'services/*/cmd'
```
returns (checksum may vary)
```json
[
   {
     "name": "service_1",
     "changed": false,
     "comment": "unmodified",
     "base": {
       "name": "service_1",
       "path": "/tmp/mono-meta/service-struct-repo.git/services/service_1/cmd/app",
       "checksum": "7b4614ff1172e7987ce3e1525216e71a",
       "reference": "refs/remotes/origin/master"
     },
     "compare": {
       "name": "service_1",
       "path": "/tmp/mono-meta/service-struct-repo.git/services/service_1/cmd/app",
       "checksum": "7b4614ff1172e7987ce3e1525216e71a",
       "reference": "refs/remotes/origin/add-new-service-and-change"
     }
   },
   {
     "name": "service_2",
     "changed": true,
     "comment": "modified",
     "base": {
       "name": "service_2",
       "path": "/tmp/mono-meta/service-struct-repo.git/services/service_2/cmd/app",
       "checksum": "1a21fcd77a2939f488d4dcf0f4a4bfc9",
       "reference": "refs/remotes/origin/master"
     },
     "compare": {
       "name": "service_2",
       "path": "/tmp/mono-meta/service-struct-repo.git/services/service_2/cmd/app",
       "checksum": "39a00e2fa982713678b683d17f87e8bb",
       "reference": "refs/remotes/origin/add-new-service-and-change"
     }
   },
   {
     "name": "service_3",
     "changed": true,
     "comment": "new",
     "base": null,
     "compare": {
       "name": "service_3",
       "path": "/tmp/mono-meta/service-struct-repo.git/services/service_3/cmd/app",
       "checksum": "50240378d271b2a7bd7fc90a999acefe",
       "reference": "refs/remotes/origin/add-new-service-and-change"
     }
   }
 ]
```

## API
Complete documentation of the API is available [here](https://godoc.org/github.com/davidae/mono-meta/mono)

### Example
```go
package main

import (
  "fmt"

  "github.com/davidae/mono-meta/mono"
)

func main() {
  repo, err := mono.NewRemote("git@github.com:davidae/service-struct-repo.git")
  if err != nil {
    panic(err)
  }

  meta := mono.NewMeta(repo, mono.Config{
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

  // cleans up cloned repo
  if err = repo.Close(); err != nil {
    panic(err)
  }
}
```
