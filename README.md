# Toolbox

![CI](https://github.com/thetreep/toolbox/workflows/CI/badge.svg?branch=master)

This repository holds a set of methods and interfaces to build production ready golang apps.

### Prerequisites

I will need _Go_, _Git_, and _Docker_

### How to build

```
    go mod vendor
    go mod download
    go build ./... # or make build
```

## Running the tests

```
    make test
```

## How to see docs

```
make doc
```

then wait 1 min and go to `http://localhost:6060/pkg/github.com?thetreep/toolbox/`

#### Application

This package holds types and methods for writing cli applications.

Example:

in `main.go`, Write:

```
package main

import (
	"log"
	"os"
	"time"

	metadata "github.com/thetreep/toolbox/application"
)

var (
	branch     string
	sha        string
	compiledAt string
	compiler   string
)

func main() {
	app, err := application.NewApplication(
		application.Config{
			Metadata: metadata.Metadata{
				Branch:     branch,
				Compiler:   compiler,
				CompiledAt: compiledAt,
				Sha:        sha,
				StartTime:  time.Now(),
			},
			Name:  "example-app",
			Usage: "usage of example app",
		})
	if err != nil {
		log.Fatalf("%v occurred with %v", err, os.Args)
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatalf("%v occurred with %v", err, os.Args)
	}
}
```

Compile with this:

```
go build  -v				        \
    -ldflags				        \
    "-X 'main.branch="branch"		\
    -X 'main.sha=commit'			\
    -X 'main.compiledAt=now'		\
    -X 'main.compiler=compiler'	\
    -s -w"					        \
    -a -installsuffix cgo main.go -o example

./example --help
```

Then you will have this:

```
NAME:
   example app - usage of example app

USAGE:
   usage of example app

VERSION:
   Branch: branch , Compiler: compiler, CompiledAt: now, Commit: commit

AUTHOR:
   the treep backend team

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --environment value    environment in which the app runs (production, staging, dev) (default: "dev") [$THETREEP_ENV]
   --help, -h             show help
   --version, -v          print the version

COPYRIGHT:
   theTreep
```
