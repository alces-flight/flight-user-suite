DIST := dist
ASSET_DIRS := opt

EXE := $(DIST)/opt/flight/usr/lib/flight-core/flight-howto

.PHONY: all assets $(EXE) clean distclean

all: assets $(EXE)

$(EXE):
	CGO_ENABLED=0 go build -o $(EXE)

assets:
	rsync -rlptgo --exclude .gitkeep $(ASSET_DIRS) $(DIST)/

clean:
	rm -f $(EXE)

distclean: clean
	rm -rf $(DIST)
