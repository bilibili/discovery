# Discovery 
[![Build Status](https://travis-ci.org/Bilibili/discovery.svg?branch=master)](https://travis-ci.org/Bilibili/discovery) 
[![Go Report Card](https://goreportcard.com/badge/github.com/Bilibili/discovery)](https://goreportcard.com/report/github.com/Bilibili/discovery)
[![codecov](https://codecov.io/gh/Bilibili/discovery/branch/master/graph/badge.svg)](https://codecov.io/gh/Bilibili/discovery)

Discovery is a based service that is production-ready and primarily used at [Bilibili](https://www.bilibili.com/) for locating services for the purpose of load balancing and failover of middle-tier servers.

## Quick Start

### env

`go1.9.x` (and later)

### build
```shell
cd $GOPATH
mkdir -p github.com/Bilibili
cd $GOPATH/github.com/Bilibili
git clone https://github.com/Bilibili/discovery.git
cd discovery/cmd/discovery
go build
```

### run
```shell
./discovery -conf discovery-example.toml -alsologtostderr
```

`-alsologtostderr` is `glog`'s flag，means print into stderr. If you hope print into file, can use `-log_dir="/tmp"`. [view glog doc](https://godoc.org/github.com/golang/glog).

### Configuration

You can view the comments in `cmd/discovery/discovery-example.toml` to understand the meaning of the config.

### Client

* [API Doc](doc/api.md)
* [Go SDK](naming/client.go) | [Example](naming/example_test.go)
* [Java SDK](https://github.com/flygit/discoveryJavaSDK)

## Intro/Arch/Practice

* [Introduction](doc/intro.md)
* [Architecture](doc/arch.md)
* [Practice in Bilibili](doc/practice.md)

## Feedback

Please report bugs, concerns, suggestions by issues, or join QQ-group 716486124 to discuss problems around source code.
