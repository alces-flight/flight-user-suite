DIST := dist
ASSET_DIRS := opt
EXE := $(DIST)/opt/flight/bin/flight

VERSION := $(shell git describe --tags --dirty --always)
COMMIT := $(shell git rev-parse HEAD)
DATE := $(shell date --rfc-3339=seconds)
LD_FLAGS := -ldflags="-X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.date=$(DATE)'"

.PHONY: all $(EXE) assets clean distclean

all: assets $(EXE)

$(EXE):
	go build -v $(LD_FLAGS) -o $(EXE)

assets:
	rsync -rlptgo $(ASSET_DIRS) $(DIST)/

clean:
	rm -f $(EXE)

distclean: clean
	rm -rf $(DIST)
