# Flight User Suite

The **Flight User Suite** provides tooling to help users make the most of their
HPC environments.

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

## Configuration

### Hooks

Hooks run code in response to events on the system. There are two types of hooks
available: "login" hooks which are automatically run when a user creates a new
login shell, and "activation" hooks that are run when the Flight environment is
activated.

The following example shows how to list available hooks, enable a hook, and list
enabled hooks.

```bash
$ sudo /opt/flight/bin/flight hooks list login
welcome
$ sudo /opt/flight/bin/flight hooks enable login welcome
Enabled welcome hook
$ sudo /opt/flight/bin/flight hooks list login --enabled
welcome
$ sudo /opt/flight/bin/flight hooks list activation
ssh-keypair-generation
$ sudo /opt/flight/bin/flight hooks enable activation ssh-keypair-generation
Enabled ssh-keypair-generation hook
$ sudo /opt/flight/bin/flight hooks list activation --enabled
ssh-keypair-generation
```

See `flight hooks --help` for more details.

### Tools

Tools are commands run by users of the Flight User Suite. Flight User Suite
includes a number of different tools, most of which are disabled by default and
can be enabled by the superuser.

The following example shows how to list available tools, enable a tool, and list
enabled tools. (Note that `flight howto` is normally enabled by default.)

```bash
$ sudo /opt/flight/bin/flight tools list
desktop
howto
$ sudo /opt/flight/bin/flight tools enable howto
Enabled flight howto tool
$ sudo /opt/flight/bin/flight tools list --enabled
howto
```

## Usage

Running `flight-start` will activate the Flight User Suite environment. The
output of this command will depend on which hooks have been configured.

With the Flight environment active, the user prompt will change and the `flight`
command will become available for both root and non-root users.

Tools can be run with `flight <toolname>`, for example:

```bash
$ flight howto --help
Usage: flight howto [--help] [list]

View user guides for your HPC environment.

    --help              Show this help message

$ flight howto list
01-about-flight-user-suite.md
```

The output of `flight --help` will list all enabled tools.

From an active Flight environment, running `flight-stop` will deactivate the
Flight User Suite and return the environment to its original settings.

### Usage detail

When a new login shell is created, the `/etc/profile.d/zz-flight-starter.sh`
file will be sourced and the login hooks run.  The following shows the output
from doing so when the `welcome` login hook is enabled.

```bash
 -[ Flight ]-
Welcome to your cluster, based on Ubuntu 22.04.5 LTS.

This cluster provides an OpenFlight HPC environment.

'flight-start' - activate Flight User Suite now
```

Once `/etc/profile.d/zz-flight-starter.sh` has been sourced, the Flight
environment can be activated by running `flight-start`.  The following example
shows the output when the `ssh-keypair-generation` activation hook is enabled.

```text
$ flight-start
                                   __ _ _       _     _  ==>
   ==>                            / _| (_)     | |   | |  ==>
  ==>   ___   _ __    ___  _ __  | |_| |_  __ _| |__ | |_  ==>
 ==>   / _ \ | '_ \  / _ \| '_ \ |  _| | |/ _` | '_ \| __|  ==>
==>   | (_) || |_) ||  __/| | | || | | | | (_| | | | | |_    ==>
 ==>   \___/ | .__/  \___||_| |_||_| |_|_|\__, |_| |_|\__|  ==>
  ==>        |_|                           __/ |           ==>
   ==>  Welcome to your cluster           |___/           ==>
    ==>  Flight User Suite v0.0.1
     ==>  Based on Ubuntu 22.04.5 LTS

Generating SSH keypair: OK
Authorizing key: OK

Your SSH config has been modified to use the generated identity file.
Please review the config if you experience issues with SSH.
SSH Config:    /home/ben/.ssh/config
Identity File: /home/ben/.ssh/id_flightcluster

Flight environment is now active.
```

# License

Eclipse Public License 2.0, see [LICENSE.txt](LICENSE.txt) for details.

Copyright (C) 2019-present Concertim Ltd.

This program and the accompanying materials are made available under
the terms of the Eclipse Public License 2.0 which is available at
[https://www.eclipse.org/legal/epl-2.0](https://www.eclipse.org/legal/epl-2.0),
or alternative license terms made available by Alces Flight Ltd -
please direct inquiries about licensing to
[licensing@alces-flight.com](mailto:licensing@alces-flight.com).

Flight User Suite is distributed in the hope that it will be
useful, but WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, EITHER
EXPRESS OR IMPLIED INCLUDING, WITHOUT LIMITATION, ANY WARRANTIES OR
CONDITIONS OF TITLE, NON-INFRINGEMENT, MERCHANTABILITY OR FITNESS FOR
A PARTICULAR PURPOSE. See the [Eclipse Public License 2.0](https://opensource.org/licenses/EPL-2.0) for more
details.
