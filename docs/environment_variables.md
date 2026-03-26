# Environment variables

Below you will find a list of environment variables used by FUS.  All of these
are exported to the calling shell.

The following are defined in `/etc/xdg/flight.config`. This is an XDG config
file, as such it can be overridden on a per-user basis. Therefore only
configuration that we're happy to allow being overridden should be included.

* `FLIGHT_ROOT` : set to the root of the FUS installation. Paths used in FUS
are either relative to this directory or absolute paths rooted at `/`. Default
`/opt/flight`.  Overriding this setting allows using an alternative
installation of FUS, which is useful for development and testing.

The following is managed by `flight-start` and `flight-stop` (that is
`${FLIGHT_ROOT}/libexec/flight-starter/main.sh`).

* `FLIGHT_ACTIVE` : set to true if the Flight environment is active in the current shell.

The following are defined by `flight-start`; they allow `flight-stop` to clean
up the shell environment when the Flight environment is deactivated. Any part
of FUS can register clean up requirements by defining `FLIGHT_ORIG_ENV_*`
variables or appending to any or all of the array variables.

* `FLIGHT_DEFINED_SYMBOLS` : array of symbols that have been defined by the
Flight environment. When the environment is exited, they are all undefined.
* `FLIGHT_ADDED_PATHS` : array of paths that the Flight environment added to
`PATH`. They are removed from `PATH` when the environment exits.
* `FLIGHT_ORIG_ENV_*` : family of environment variables each containing the
original value of the environment variable as it was when the Flight
environment was activated. E.g., `FLIGHT_ORIG_ENV_PS1` contains the original
setting for `PS1`.  Any variables registered here will be reset when the Flight
environment exits.
* `FLIGHT_ON_EXIT` : array of functions to evaluate when the Flight environment
is exited. This allows handling more complex clean up routines that are not
easily managed with the above environment variables.

The following are defined in `${FLIGHT_ROOT}/etc/flight-starter.config` and support branding the FUS.

* `FLIGHT_STARTER_CLUSTER_NAME` : name of the cluster. Used in the banner and in the PS1 prompt configuration.
* `FLIGHT_STARTER_DESC` : description of the Flight environment used in the login welcome hook.
* `FLIGHT_STARTER_PRODUCT` : brand name of the FUS product.
* `FLIGHT_STARTER_RELEASE` : FUS release (version number) e.g., v0.0.1.

## Non-environment variables

FUS defines other variables that are not exported to the calling shell.  Their
use is encapsulated within a single process or function.  Some of them are
listed below.

* `FLIGHT_SSH_LOWEST_UID`, `FLIGHT_SSH_SKIP_USERS`, `FLIGHT_SSH_KEYNAME`,
`FLIGHT_SSH_DIR`, `FLIGHT_SSH_LOG` : defined in
`${flight_root}/etc/ssh-keypair-generation.config` and used to configure the
generation of a password-less SSH key.
