---
admin: true
---
# Managing Hooks and Tools in Flight User Suite

The Flight User Suite contains multiple components which can be enabled to 
expose them to end-users and tailor the Flight User Suite tool to your needs.

## Available Hooks and Tools

### Hooks

Hooks run code in response to events on the system. There are two types of hooks
available: "login" hooks which are automatically run when a user creates a new
login shell, and "activation" hooks that are run when the Flight environment is
activated.

**welcome**
Print a short welcome message upon terminal login to the HPC environment 
describing how a user can activate the Flight environment.

**ssh-keypair-generation**
Generate a trusted ssh-keypair for the user when the Flight environment is 
activated. This key is named `id_flightcluster` and is added to the user's 
`~/.ssh/authorized_keys` file as well as being automatically used with SSH 
commands through `~/.ssh/config`.
In HPC environments with networked storage providing user home directories this
will quickly enable passwordless SSH between hosts.

**user-symlinks**
Generate a symbolic link for the user when the Flight environment is activated. 
By default this will create a link `~/scratch` to `/scratch/USERNAME`. For more
information on this hook see the "Managed Cluster Admin Guide" in `howto` when
the hook has been enabled. 

### Tools 

Tools are commands run by users of the Flight User Suite. Flight User Suite
includes a number of different tools, most of which are disabled by default and
can be enabled by the superuser.

**howto**
An interface for Flight User Suite documentation. Enabled by default.

**desktop**
Create and manage virtual desktop sessions across multiple hosts.

**mfa**
Manages multi-factor authentication configuration for end users. Additional 
system integration is required and is outlined in "Admin Guides > Flight MFA"
as part of `howto` when the tool has been enabled.

## Configuration

### Hooks

The following example shows how to list available hooks, enable a hook, and list
enabled hooks.

```bash
$ sudo /opt/flight/bin/flight hooks list
┌────────────┬────────────────────────┬─────────────┐
│   Event    │          Name          │   Enabled   │
├────────────┼────────────────────────┼─────────────┤
│ login      │ welcome                │ ❌ Disabled │
│ activation │ ssh-keypair-generation │ ❌ Disabled │
└────────────┴────────────────────────┴─────────────┘
$ sudo /opt/flight/bin/flight hooks enable login welcome
Enabled welcome hook
$ sudo /opt/flight/bin/flight hooks list --enabled
┌───────┬─────────┬─────────────┐
│ Event │  Name   │   Enabled   │
├───────┼─────────┼─────────────┤
│ login │ welcome │ ✅ Enabled  │
└───────┴─────────┴─────────────┘
$ sudo /opt/flight/bin/flight hooks enable activation ssh-keypair-generation
Enabled ssh-keypair-generation hook
$ sudo /opt/flight/bin/flight hooks list
┌────────────┬────────────────────────┬─────────────┐
│   Event    │          Name          │   Enabled   │
├────────────┼────────────────────────┼─────────────┤
│ login      │ welcome                │ ✅ Enabled  │
│ activation │ ssh-keypair-generation │ ✅ Enabled  │
└────────────┴────────────────────────┴─────────────┘
```

See `flight hooks --help` for more details.

### Tools

The following example shows how to list available tools, enable a tool, and list
enabled tools. (Note that `flight howto` is normally enabled by default.)

```bash
$ sudo /opt/flight/bin/flight tools list
┌─────────┬─────────────┐
│  Name   │   Enabled   │
├─────────┼─────────────┤
│ desktop │ ❌ Disabled │
│ howto   │ ✅ Enabled  │
└─────────┴─────────────┘
$ sudo /opt/flight/bin/flight tools enable desktop
Enabled flight desktop tool
$ sudo /opt/flight/bin/flight tools list --enabled
┌─────────┬─────────────┐
│  Name   │   Enabled   │
├─────────┼─────────────┤
│ desktop │ ✅ Enabled  │
│ howto   │ ✅ Enabled  │
└─────────┴─────────────┘
```
