SHELL := /bin/bash
DIST := dist
OPT := opt
ASSET_DIRS := assets views static
EXE := $(DIST)/opt/flight/usr/libexec/web-suite/service
FLIGHT_ROOT := $(DIST)/opt/flight
FLIGHT_STATE_ROOT := tmp
WEB_ROOT := $(FLIGHT_ROOT)/var/web-suite
TAILWIND := ./bin/tailwind/tailwindcss
GOTESTSUM := $(shell type -p gotestsum)
ESBUILD := ./node_modules/.bin/esbuild

VERSION := $(shell git describe --tags --dirty --always)
COMMIT := $(shell git rev-parse HEAD)
DATE := $(shell date --rfc-3339=seconds)
LD_FLAGS := -ldflags="-X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.date=$(DATE)'"

.PHONY: all $(EXE) assets build-css build-javascript clean distclean dev help run test

all: $(EXE) assets

$(EXE):
	CGO_ENABLED=0 go build -v $(LD_FLAGS) -o $(EXE)

assets: build-css build-javascript
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

$(ESBUILD):
	npm install

build-javascript: $(ESBUILD)
	$(ESBUILD) \
		assets/javascript/application.js \
		--bundle \
		--outfile=static/js/application.js \
		--minify

dev: $(TAILWIND) $(ESBUILD)
	FLIGHT_ROOT=$(OPT)/flight FLIGHT_STATE_ROOT=$(FLIGHT_STATE_ROOT) air

run: all
	FLIGHT_ROOT=$(FLIGHT_ROOT) FLIGHT_STATE_ROOT=$(FLIGHT_STATE_ROOT) $(EXE)

test:
ifdef GOTESTSUM
	FLIGHT_ROOT=$(OPT)/flight FLIGHT_STATE_ROOT=$(FLIGHT_STATE_ROOT) $(GOTESTSUM) -- ./... -count 1
else
	FLIGHT_ROOT=$(OPT)/flight FLIGHT_STATE_ROOT=$(FLIGHT_STATE_ROOT) go test ./... -count 1
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
