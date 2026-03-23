DIST := dist
ASSET_DIRS := opt

.PHONY: all assets clean distclean

all: assets

assets:
	rsync -rlptgo $(ASSET_DIRS) $(DIST)/

clean:

distclean: clean
	rm -rf $(DIST)
