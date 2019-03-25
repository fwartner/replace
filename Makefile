SOURCE = $(wildcard *.go)
TAG ?= $(shell git describe --tags)
GOBUILD = go build -ldflags '-w'

ALL = \
	$(foreach arch,64 32,\
	$(foreach suffix,linux osx win.exe,\
		build/gr-$(arch)-$(suffix))) \
	$(foreach arch,arm arm64,\
		build/gr-$(arch)-linux)

all: build

build: clean $(ALL)

clean:
	rm -f $(ALL)

win.exe = windows
osx = darwin
build/gr-64-%: $(SOURCE)
	@mkdir -p $(@D)
	CGO_ENABLED=0 GOOS=$(firstword $($*) $*) GOARCH=amd64 $(GOBUILD) -o $@

build/gr-32-%: $(SOURCE)
	@mkdir -p $(@D)
	CGO_ENABLED=0 GOOS=$(firstword $($*) $*) GOARCH=386 $(GOBUILD) -o $@

build/gr-arm-linux: $(SOURCE)
	@mkdir -p $(@D)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 $(GOBUILD) -o $@

build/gr-arm64-linux: $(SOURCE)
	@mkdir -p $(@D)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) -o $@

release: build
	github-release release -u fwartner -r replace -t "1.0.0" -n "1.0.0" --description "1.0.0"
	@for x in $(ALL); do \
		echo "Uploading $$x" && \
		github-release upload -u fwartner \
                              -r replace \
                              -t 1.0.0 \
                              -f "$$x" \
                              -n "$$(basename $$x)"; \
	done
