DIST := dist
ASSET_DIRS := etc opt

.PHONY: all assets clean distclean

all: assets

assets:
	rsync -rlptgo --exclude .gitkeep $(ASSET_DIRS) $(DIST)/

clean:

distclean: clean
	rm -rf $(DIST)
