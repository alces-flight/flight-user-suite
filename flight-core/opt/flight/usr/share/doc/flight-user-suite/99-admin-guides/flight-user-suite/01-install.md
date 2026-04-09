---
admin: true
---
## Installation

1. Download the latest release from [the releases page](https://github.com/concertim/flight-user-suite/releases).
2. Copy the release tarball to the target system.
3. Untar the release to `/`: it will install Flight User Suite to `/opt/flight`
   along with configuration files `/etc/xdg/flight.config` and
   `/etc/profile.d/zz-flight-starter.sh`.

Once Flight User Suite has been installed, you can enable the hooks and tooling
that you wish to make available.

### Upgrading

Upgrading is **not** supported, but is possible by unpacking a newer release
over the top of an existing installation. Configuration of hooks and tools may
be lost in this process and `/etc/xdg/flight.config` will be overwritten.
