DIST := dist
ASSET_DIRS := assets views static
EXE := $(DIST)/opt/flight/usr/libexec/web-suite/service

VERSION := $(shell git describe --tags --dirty --always)
COMMIT := $(shell git rev-parse HEAD)
DATE := $(shell date --rfc-3339=seconds)
LD_FLAGS := -ldflags="-X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.date=$(DATE)'"

.PHONY: all $(EXE) clean distclean

all: $(EXE)

$(EXE): assets
	CGO_ENABLED=0 go build -v $(LD_FLAGS) -o $(EXE)

assets: build-css
	mkdir -p $(DIST)/opt/flight/var/web-suite/
	rsync -rlptgo --exclude .gitkeep $(ASSET_DIRS) $(DIST)/opt/flight/var/web-suite/

build-css:
	./bin/tailwind/tailwindcss \
		--config ./tailwind.config.js \
		-i ./assets/styles/application.css \
		-o ./static/styles.css \

clean:
	rm -f $(EXE)

distclean: clean
	rm -rf $(DIST)
