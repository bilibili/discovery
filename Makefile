# Go parameters
GOCMD=GO111MODULE=on CGO_ENABLED=0 go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

all: test build
build:
	rm -rf dist/
	mkdir -p dist/conf
	cp cmd/discovery/discovery-example.toml dist/conf/discovery.toml
	$(GOBUILD) -o dist/bin/discovery cmd/discovery/main.go

test:
	$(GOTEST) -v ./...

clean:
	rm -rf dist/

run:
	nohup dist/bin/discovery -conf dist/conf -confkey discovery.toml -log.dir dist/log & > dist/nohup.out

stop:
	pkill -f dist/bin/discovery
