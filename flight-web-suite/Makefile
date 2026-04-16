DIST := dist
EXE := $(DIST)/opt/flight/usr/libexec/web-suite/service

VERSION := $(shell git describe --tags --dirty --always)
COMMIT := $(shell git rev-parse HEAD)
DATE := $(shell date --rfc-3339=seconds)
LD_FLAGS := -ldflags="-X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.date=$(DATE)'"

.PHONY: all $(EXE) clean distclean

all: $(EXE)

$(EXE): build-css
	CGO_ENABLED=0 go build -v $(LD_FLAGS) -o $(EXE)

clean:
	rm -f $(EXE)

distclean: clean
	rm -rf $(DIST)

build-css:
	./bin/tailwind/tailwindcss \
        --config ./tailwind.config.js \
	-i ./assets/styles/application.css \
	-o ./static/styles.css \
