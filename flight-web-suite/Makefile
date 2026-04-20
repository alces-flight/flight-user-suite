SHELL := /bin/bash
DIST := dist
OPT := opt
ASSET_DIRS := assets views static
EXE := $(DIST)/opt/flight/usr/libexec/web-suite/service
FLIGHT_ROOT := $(DIST)/opt/flight
WEB_ROOT := $(FLIGHT_ROOT)/var/web-suite
TAILWIND := ./bin/tailwind/tailwindcss
GOTESTSUM := $(shell type -p gotestsum)

VERSION := $(shell git describe --tags --dirty --always)
COMMIT := $(shell git rev-parse HEAD)
DATE := $(shell date --rfc-3339=seconds)
LD_FLAGS := -ldflags="-X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.date=$(DATE)'"

.PHONY: all $(EXE) assets build-css clean distclean dev help run test

all: $(EXE) assets

$(EXE):
	CGO_ENABLED=0 go build -v $(LD_FLAGS) -o $(EXE)

assets: build-css
	mkdir -p $(WEB_ROOT)
	rsync -rlptgo --exclude .gitkeep $(ASSET_DIRS) $(WEB_ROOT)/
	rsync -rlptgo --exclude .gitkeep $(OPT) $(DIST)/

$(TAILWIND):
	./bin/tailwind/download-tailwindcss.sh

build-css: $(TAILWIND)
	$(TAILWIND) \
		--config ./tailwind.config.js \
		-i ./assets/styles/application.css \
		-o ./static/styles.css

dev: $(TAILWIND)
	air

run: all
	FLIGHT_ROOT=$(FLIGHT_ROOT) $(EXE)

test:
ifdef GOTESTSUM
	$(GOTESTSUM) -- ./... -count 1
else
	go test ./... -count 1
endif

help:
	@printf '%s\n' \
		'all      Build the service and package runtime assets into $(DIST)' \
		'assets   Stage packaged assets into $(WEB_ROOT)' \
		'dev      Run server in "development" mode' \
		'run      Build and launch the packaged service locally' \
		'test     Run tests' \
		'clean    Remove generated binaries' \
		'distclean Remove $(DIST)'

clean:
	rm -f $(EXE)

distclean: clean
	rm -rf $(DIST)
