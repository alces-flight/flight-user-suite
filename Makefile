DIST := dist
MODULES := flight-starter flight-core flight-desktop flight-howto ssh-keypair-generation
VERSION := $(shell git describe --tags --dirty --always)
KERNEL := $(shell uname -s)
ARCH := $(shell uname -p)
TARFILE := flight-user-suite_$(VERSION)_$(KERNEL)_$(ARCH).tar.gz

.PHONY: all clean distclean $(MODULES) $(DIST)

all: $(TARFILE)

$(MODULES):
	$(MAKE) -C $@

$(DIST): $(MODULES)
	rsync -rlptgo $(foreach module,$(MODULES),$(module)/dist/) $(DIST)/
# Disable all tools by default except flight-howto
	chmod 0444 $(DIST)/opt/flight/usr/lib/flight-core/*
	chmod 0555 $(DIST)/opt/flight/usr/lib/flight-core/flight-howto
# Note: if we ever want default-enabled tools that have documentation we will
# also need to symlink the docs into
# $(DIST)/opt/flight/usr/share/doc/howtos-enabled here.
	cp README.md $(DIST)/opt/flight/usr/share/doc/flight-user-suite/01-about-flight-user-suite.md
	cp LICENSE.txt $(DIST)/opt/flight/usr/share/doc/flight-user-suite/

$(TARFILE): $(DIST)
	tar czf $@ --owner=root:0 --group=root:0 -C $(DIST) .

clean:
	rm -f flight-user-suite*.tar.gz
	for m in $(MODULES) ; do \
		$(MAKE) -C $$m clean ; \
	done;

distclean: clean
	rm -rf $(DIST)
	for m in $(MODULES) ; do \
		$(MAKE) -C $$m distclean ; \
	done;
