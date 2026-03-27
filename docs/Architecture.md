# Architecture

## Flight starter

Flight Starter is responsible for activating the Flight environment in a shell.
It modifies the shell's `PATH` and `PS1` and sets a number of environment
variables.

It installs two files outside of `/opt/flight`: `/etc/xdg/flight.config` and
`/etc/profile.d/zz-flight-starter.sh`.

When `/etc/profile.d/zz-flight-starter.sh` is sourced the following happens:

* The `FLIGHT_ROOT` environment variable is read from `/etc/xdg/flight.config`.
* Any enabled "login" hooks are run (see below for more details).
* A `flight-start` function is defined which activates the Flight environment.

When `flight-start` is called the following happens:

* Any scripts in `${FLIGHT_ROOT}/etc/profile.d/` are sourced.
* Any enabled "activation" hooks are run (see below for more details).
* `${FLIGHT_ROOT}/bin/` is added to the shell's `PATH`.
* A `flight-stop` function is defined which deactivates the Flight environment.
* The `flight-start` function is undefined.

When `flight-stop` is called any changes to the shell's environment are
undone. This includes:

* Removing `${FLIGHT_ROOT}/bin/` from the shell's `PATH`.
* Redefining the `flight-start` function.
* Undefining `flight-stop` function.
* Undoing any changes made by scripts in `${FLIGHT_ROOT}/etc/profile.d/`, such as modifying the `PS1`.

## Flight core

Flight core is the main entry point for configuring FUS and for accessing the
tools it provides.  It is available as a binary at `${FLIGHT_ROOT}/bin/flight`.
It can be ran using its full path without activating the Flight environment, or
simply as `flight` once the Flight environment is activated.

For root it allows managing the availability of tools and hooks and for
accessing any enabled tools.

For non-root users it allows accessing any enabled tools.

### Limitation

Currently, if a tool or hook is enabled it is enabled for both root and
non-root users.  It is expected that later work will allow for root-only tools
and hooks.

## Hooks

There are two types of hooks available: "login" hooks which are automatically
run when a user creates a new login shell, and "activation" hooks that are run
when the Flight environment is activated.

Login hooks are defined as any file matching the glob
`${FLIGHT_ROOT}/usr/lib/hooks/login/*`, whilst activation hooks are defined as
any file matching the glob `${FLIGHT_ROOT}/usr/lib/hooks/activation/*`.

A hook is enabled if it is executable and is disabled otherwise.

When a login shell is created all enabled login hooks will be executed, this is
done as a result of login shells sourcing
`/etc/profile.d/zz-flight-starter.sh`.

When `flight-start` is ran, all enabled activation hooks will be executed.

### Hook limitations

The current implementation of hooks requires them to be executed---it does not
support sourcing them.  Executing hooks prevents them from modifying the
shell's environment, which prevents implementing all of the
`${FLIGHT_ROOT}/etc/profile.d/*` scripts as hooks. An example is `02-prompt.sh`
which needs to modify the shell's `PS1` variable.

There is currently no support for disabling any scripts in
`${FLIGHT_ROOT}/etc/profile.d/*`.  They are always ran when the Flight
environment is activated.

It is expected that future work will allow hooks to be sourced whilst still
supporting enabling and disabling of any hook whether executed or sourced.

## Tools

Tools are programs that a user or admin can run to achieve a specific task.
Unlike hooks, which run automatically, tools are only run on demand.  An
example would be `flight desktop` that allows for managing virtual desktop
environments.

Tools are defined as any file matching the glob
`${FLIGHT_ROOT}/usr/lib/flight-core/flight-*`. A tool is enabled if it is
executable and is disabled otherwise.

An enabled tool can be run as `${FLIGHT_ROOT}/bin/flight <tool>`. E.g., running
`${FLIGHT_ROOT}/bin/flight desktop --help` would run
`${FLIGHT_ROOT}/usr/lib/flight-core/flight-desktop --help`. If the Flight
environment is activated, this can be simplified to `flight <tool>`.

### Howto guides

A tool can provide howto guides by creating markdown files in
`${FLIGHT_ROOT}/usr/share/doc/flight-<tool>/`.

When a tool is enabled these are symlinked into
`${FLIGHT_ROOT}/usr/share/doc/howtos-enabled/` which makes them available to
the the `flight-howto` tool.

Similarly, when a tool is disabled, those symlinks are removed.

## Packaging

There are separate directories (aka subprojects) for `flight-starter`,
`flight-core` and the various tools.  Hooks might be defined in
`flight-starter` one of the tool subprojects or as their own subproject.

This separation allows for each subproject to be developed separately and to
manage its own build and packaging requirements without conflict.

Each subproject is required to contain a `Makefile` defining an `all` target.
This target should build the subproject to `<subproject>/dist/`. The layout
inside `<subproject>/dist/` should match the desired layout after installation,
e.g., if subproject requires that a file is installed at
`/opt/flight/usr/lib/flight-core/flight-desktop`, the `all` target should copy
that file to `<subproject>/dist/opt/flight/usr/lib/flight-core/flight-desktop`.

A top-level Makefile builds all of the subprojects and creates a tarball from
them suitable for extration over `/`.
