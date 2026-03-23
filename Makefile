DIST := dist
MODULES := flight-starter flight-core flight-desktop flight-howto
VERSION := $(shell git describe --tags --dirty --always)
KERNEL := $(shell uname -s)
ARCH := $(shell uname -p)
TARFILE := flight-user-suite-$(VERSION)-$(KERNEL)-$(ARCH).tar.gz

.PHONY: all clean distclean $(MODULES) $(DIST)

all: $(TARFILE)

$(MODULES):
	$(MAKE) -C $@

$(DIST): $(MODULES)
	rsync -rlptgo $(foreach module,$(MODULES),$(module)/dist/) $(DIST)/

$(TARFILE): $(DIST)
	tar czf $@ -C $(DIST) .

clean:
	rm -f flight-user-suite-*.tar.gz
	for m in $(MODULES) ; do \
		$(MAKE) -C $$m clean ; \
	done;

distclean: clean
	rm -rf $(DIST)
	for m in $(MODULES) ; do \
		$(MAKE) -C $$m distclean ; \
	done;
