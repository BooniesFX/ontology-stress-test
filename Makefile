GOFMT=gofmt
GC=go build
VERSION := $(shell git describe --abbrev=4 --dirty --always --tags)
Minversion := $(shell date)
BUILD_NODE_PAR = -ldflags "-X github.com/ontio/ontology-stress-test/common/config.Version=$(VERSION)" #-race
BUILD_NODECTL_PAR = -ldflags "-X main.Version=$(VERSION)"

net-bench:
	$(GC)  $(BUILD_NODE_PAR) -o net-stress-test main.go
	$(GC)  $(BUILD_NODECTL_PAR) testcli.go
bench:
	$(GC)  $(BUILD_NODE_PAR) -o ont-bench ontbench.go

format:
	$(GOFMT) -w main.go

clean:
	rm -rf *.8 *.o *.out *.6
