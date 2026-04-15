DIST := dist
ASSET_DIRS := opt
TOOL_EXE := $(DIST)/opt/flight/usr/lib/flight-core/flight-web
SERVICE_EXE := $(DIST)/opt/flight/usr/libexec/web-suite/service

VERSION := $(shell git describe --tags --dirty --always)
COMMIT := $(shell git rev-parse HEAD)
DATE := $(shell date --rfc-3339=seconds)
LD_FLAGS := -ldflags="-X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.date=$(DATE)'"

.PHONY: all $(TOOL_EXE) $(SERVICE_EXE) assets clean distclean

all: assets $(TOOL_EXE) $(SERVICE_EXE)

$(TOOL_EXE):
	CGO_ENABLED=0 go build -C tool -v $(LD_FLAGS) -o ../$(TOOL_EXE)

$(SERVICE_EXE):
	CGO_ENABLED=0 go build -C service -v $(LD_FLAGS) -o ../$(SERVICE_EXE)

assets:
	rsync -rlptgo --exclude .gitkeep $(ASSET_DIRS) $(DIST)/

clean:
	rm -f $(EXE)

distclean: clean
	rm -rf $(DIST)
