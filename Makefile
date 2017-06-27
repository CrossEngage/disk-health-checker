APPNAME   := disk-health-checker
DIST_DIR  := ./dist
PLATFORMS := linux-386 linux-amd64 linux-arm

RELEASE   := $(shell git describe --all --always)

build:
	go generate
	go get -v -t ./...
	go test -v ./...
	go build -v


dist: $(PLATFORMS)
$(PLATFORMS):
	$(eval GOOS := $(firstword $(subst -, ,$@)))
	$(eval GOARCH := $(lastword $(subst -, ,$@)))
	mkdir -p $(DIST_DIR)
	env GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(DIST_DIR)/$(APPNAME).$(RELEASE).$@


release: dist
	go get github.com/tcnksm/ghr
	if [ "x$$(git config --global --get github.token)" = "x" ]; then echo "Missing github.token in your git config"; fi
	ghr -recreate -u crossengage $(RELEASE) dist/
