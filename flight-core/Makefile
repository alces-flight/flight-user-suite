DIST := dist
ASSET_DIRS := opt
EXE := $(DIST)/opt/flight/bin/flight
FLIGHT_ROOT := $(DIST)/opt/flight
FLIGHT_STATE_ROOT := $(DIST)/opt/flight/var/lib

VERSION := $(shell git describe --tags --dirty --always)
COMMIT := $(shell git rev-parse HEAD)
DATE := $(shell date --rfc-3339=seconds)
LD_FLAGS := -ldflags="-X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.date=$(DATE)'"

.PHONY: all $(EXE) assets clean distclean shell_completions

all: assets $(EXE) shell_completions

$(EXE):
	CGO_ENABLED=0 go build -v $(LD_FLAGS) -o $(EXE)

assets:
	rsync -rlptgo --exclude .gitkeep $(ASSET_DIRS) $(DIST)/

shell_completions: $(DIST)/usr/share/bash-completion/completions/flight

$(DIST)/usr/share/bash-completion/completions/flight: $(EXE)
	mkdir -p $(dir $@)
	FLIGHT_ROOT=$(FLIGHT_ROOT) FLIGHT_STATE_ROOT=$(FLIGHT_STATE_ROOT) $(EXE) completion bash > $@

clean:
	rm -f $(EXE)

distclean: clean
	rm -rf $(DIST)
