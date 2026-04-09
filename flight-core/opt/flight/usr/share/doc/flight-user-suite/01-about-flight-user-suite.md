# About Flight User Suite

The Flight User Suite provides tooling to help you get the most out of your
HPC environment.

## Usage

Running `flight-start` will activate the Flight User Suite environment. 

With the Flight environment active, your prompt will change and the `flight`
command will become available.

Tools can be run with `flight <toolname>`, for example:

```bash
$ flight howto --help
NAME:
   flight howto - View user guides for your HPC environment

USAGE:
   flight howto [global options] [command [command options]]

DESCRIPTION:
   View user guides for your HPC environment

COMMANDS:
   list, l, ls  List available user guides
   show, s      Open a user guide for viewing in the terminal
   help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help

COPYRIGHT:
   (c) 2026 Stephen F Norledge & Concertim Ltd.
```

The output of `flight --help` will list all enabled tools.

From an active Flight environment, running `flight-stop` will deactivate the
Flight User Suite and return the environment to its original settings.

## Configuring Your Environment

Using `flight config` you can set the behaviour of your Flight environment.

Currently supported configuration options:
- `autostart`: This controls whether the Flight User Suite environment is
  activated by default on login. Valid options are `on` and `off`. 

To set a configuration option:
```
flight config set OPTION VALUE
```

## Admin Help

If you're the admin of the system you can view administrative documentation by 
running `flight howto list` in an active flight environment as the `root` user.
