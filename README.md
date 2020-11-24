# Discovery 
[![Build Status](https://travis-ci.org/bilibili/discovery.svg?branch=master)](https://travis-ci.org/bilibili/discovery) 
[![Go Report Card](https://goreportcard.com/badge/github.com/bilibili/discovery)](https://goreportcard.com/report/github.com/bilibili/discovery)
[![codecov](https://codecov.io/gh/Bilibili/discovery/branch/master/graph/badge.svg)](https://codecov.io/gh/Bilibili/discovery)

Discovery is a based service that is production-ready and primarily used at [Bilibili](https://www.bilibili.com/) for locating services for the purpose of load balancing and failover of middle-tier servers.

## Quick Start

### env

`go1.12.x` (and later)

### build
```shell
cd $GOPATH/src
git clone https://github.com/bilibili/discovery.git
cd discovery/cmd/discovery
go build
```

### run
```shell
./discovery -conf discovery.toml -alsologtostderr
```

`-alsologtostderr` is `glog`'s flag，means print into stderr. If you hope print into file, can use `-log.dir="/tmp"`. [view glog doc](https://godoc.org/github.com/golang/glog).

### Configuration

You can view the comments in `cmd/discovery/discovery.toml` to understand the meaning of the config.

### Client

* [API Doc](doc/api.md)
* [Go SDK](naming/client.go) | [Example](naming/example_test.go)
* [Java SDK](https://github.com/flygit/discoveryJavaSDK)
* [CPP SDK](https://github.com/brpc/brpc/blob/master/src/brpc/policy/discovery_naming_service.cpp)
* [Python SDK](https://github.com/tomwei7/discovery-client)
* [other language](doc/sdk.md)

## Intro/Arch/Practice

* [Introduction](doc/intro.md)
* [Architecture](doc/arch.md)
* [Practice in Bilibili](doc/practice.md)

## Feedback

Please report bugs, concerns, suggestions by issues, or join QQ-group 716486124 to discuss problems around source code.
