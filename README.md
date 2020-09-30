# Toolbox

![Build and Test Go](https://github.com/thetreep/toolbox/workflows/Build%20and%20test%20Go/badge.svg)

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
    make test # or go ./...
```

## How to see docs

```
make doc
```

then wait 1 min and go to `http://localhost:6060/pkg/github.com?thetreep/toolbox/`
