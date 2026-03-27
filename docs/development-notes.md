# Tooling, e.g., desktop and howto

## Main tool / entry point

Provide a single entry point to all other tools at `/opt/flight/bin/flight`.

Running this executable provides a list of available commands, e.g.,

```
$ /opt/flight/bin/flight
Usage: flight [--version] [--help] <command> [<args>]

The following commands are available:

    desktop    Manage virtual desktop environments
    howto      Get help on using the flight user suite or your HPC cluster
```

Running with a command runs requested tool, e.g.,

```
$ /opt/flight/bin/flight howto --help
Usage: flight howto [--version] [--help] <command> [<args>]

The following commands are available:

    list      List available howto guides
    show      View a howto guide
    search    Search for a howto guide
```


## Individual tools

Shamelessly ~ripped off of~ borrowed from Git's plugin architecture.

* Executable file installed into `/opt/flight/usr/lib/flight-core/`.
* Running `flight` / `flight --help` lists executable tools in `/opt/flight/usr/lib/flight-core/`.
* Running, say, `flight howto` executes `/opt/flight/usr/lib/flight-core/flight-howto` passing any and all arguments to it.

This setup has the following features.

* Allows tools to be developed independently in the most appropriate language.
* Allows non-standard cluster-specific tooling to be developed and easily integrated.


## Toggle on-and-off-able

Provide a command to toggle tool availability, e.g., `flight tools desktop
<on|off>`. This command:

1. toggles the executable permission for the corresponding tool in `/opt/flight/usr/lib/flight-core/`.
2. adjusts symlinks in `/opt/flight/usr/share/doc/howtos-enabled/`.

If the tool is not executable:

* It does not appear in `flight --help` output.
* Running it via `flight <tool>` will fail with command not known.
* Running it directly via `/opt/flight/usr/lib/flight-core/flight-<tool>` will fail with file not executable.

If the tool's howto guides are not symlinked into `/opt/flight/usr/share/doc/howtos-enabled/`:

* `flight-howto` will be unaware of them.

Issues with this approach:

* Tools are discoverable via standard CLI tools such as `ls` and `find`.
* Tools can be enabled by anyone with root access.
* Tool availability can diverge from one machine to another (this could potentially be a feature).

## Disk layout

Putting all of that together, we would have the following on-disk layout.

```
/opt/flight/bin/flight
/opt/flight/usr/lib/flight-core/flight-desktop
/opt/flight/usr/lib/flight-core/flight-howto
/opt/flight/usr/share/doc/howtos-enabled/
/opt/flight/usr/share/doc/flight-core/01-access-your-cluster.md
/opt/flight/usr/share/doc/flight-core/02-submit-a-job.md
/opt/flight/usr/share/doc/flight-howto/01-use-flight-howto.md
/opt/flight/usr/share/doc/flight-desktop/01-launch-desktop.md
```


## Distribution

Could be via a `tar.gz` file with the above layout.  Could be via an RPM or a
DEB with the above layout.

In any case, there is no post-installation installation of tooling.


# Integration

By default, `/opt/flight/bin` is not on `PATH`, so installation is opt-in.

We add a profile script at `/etc/profile.d/flight.sh`. This script defines a
Bash function `flight-start`. To opt-in a user runs `flight-start` which does
the following:

* adds `/opt/flight/bin` to `PATH`
* sources profile scripts at `/opt/flight/etc/profile.d/`

The following scripts in `/opt/flight/etc/profile.d/` are sourced:

* `banner.sh` - displays a welcome to "Flight User Suite" banner.
* `prompt.sh` - modifies `PS1` to include that the flight environment is active.
* `stop.sh` - un-defines `flight-start` and defines `flight-stop`. 

When `flight-stop` is ran it 

1. re-defines `flight-start`
2. un-defines `flight-stop`
3. removes the side effects of running `flight-start`. This requires we track
   the initial `PATH` and `PS1`.


## Disk layout

Disk layout for integration / starter work:

```
/etc/profile.d/flight.sh
/opt/flight/etc/profile.d/01-banner.sh
/opt/flight/etc/profile.d/01-prompt.sh
/opt/flight/etc/profile.d/01-stop.sh
```

## Future enhancements

### Support shells other than Bash.

To support, say, tcsh we would add the following files:

```
/etc/profile.d/flight.csh
/opt/flight/etc/profile.d/01-banner.csh
/opt/flight/etc/profile.d/01-prompt.csh
/opt/flight/etc/profile.d/01-stop.csh
```

The `flight-start` function defined in `/etc/profile.d/flight.csh` would source
the `/opt/flight/etc/profile.d/*.csh` profiles scripts.


### Add support for automatically starting the flight environment

Add support for automatically starting the flight environment when a user logs
in.  It should be possible to set a cluster-wide default and to override the
default on a per-user basis.

```
$ flight config autostart <on|off> [--global]
```

Possible implementation:

With `--global` present edit `/etc/xdg/flight.conf`, without `--global` edit
`~/.config/flight.conf`. In either case, adjust an `autostart=True` /
`autostart=False` line as appropriate.


# More thoughts

## Built-in vs plugin

The Flight Core executable, aka `/opt/flight/bin/flight`, should probably have
the `tools` command as a builtin to prevent disabling the `tools` tool.

Both `howto` and `desktop` are implemented as plugins. `config` is probably a
plugin.


## Building the distribution

We want to have a single file to distribute, e.g., a single `*.tar.gz` or a
single RPM.  The easiest way to build that file is likely to have a mono-repo.

We can then have a `build.sh`/`package.sh` script that has access to all the
files it needs in the repo. This would make running it on CI easy.

## Parallel development

I see the following avenues for parallel development.

* Flight Integration - Bash scripts to modify `PATH` and `PS1`.
* Flight Core (a.k.a., `/opt/flight/bin/flight`) - wrapper to launch other scripts, plus `tools` command to manage tool availability.
* `flight-desktop` - standalone tool to start/stop/list virtual desktops.
* `flight-howto` - standalone tool to list/show howto documents in a specific directory.


# Areas lacking thought

* Exact functionality of each tool.
* How flight desktop types will work.
