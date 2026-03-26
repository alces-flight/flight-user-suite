DIST := dist
ASSET_DIRS := etc opt

VERSION := $(shell git describe --tags --dirty --always)

.PHONY: all assets clean distclean

all: assets $(DIST)/opt/flight/etc/flight-starter.config

assets:
	rsync -rlptgo --exclude .gitkeep \
		--exclude opt/flight/etc/flight-starter.config \
		$(ASSET_DIRS) $(DIST)/

$(DIST)/opt/flight/etc/flight-starter.config: opt/flight/etc/flight-starter.config
	sed 's/^\(FLIGHT_STARTER_RELEASE=\).*/\1$(VERSION)/' $< > $@

clean:

distclean: clean
	rm -rf $(DIST)
