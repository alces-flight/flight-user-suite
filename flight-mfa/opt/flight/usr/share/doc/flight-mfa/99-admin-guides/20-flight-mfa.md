---
admin: true
name: "Admin Guides > Flight MFA"
---
# Flight MFA

Flight MFA adds a user-facing command for generating and displaying a
time-based one-time password (TOTP) secret. When it is combined with SSH
and PAM configuration, users can be required to authenticate with a
password and a changing one-time code in addition to their SSH key.

Multi-factor authentication (MFA) means a user must prove their identity
with more than one factor. In this setup the factors are typically:

- something the user has: an authenticator application that can generate
  TOTP codes
- something the user knows: their account password
- something the user has already configured for SSH: their private key

MFA reduces the impact of stolen credentials. A compromised SSH private
key or password is no longer enough on its own to complete a login. This
is particularly useful for shared clusters, systems exposed to less
trusted networks, and environments where administrators want stronger
assurance for interactive logins.

## What Flight MFA Provides

The `flight mfa` command allows users to:

- generate a TOTP secret with `flight mfa generate`
- display an existing TOTP secret with `flight mfa show`

The command stores the generated secret in the user account, by default
at:

```text
~/.config/flight/mfa.dat
```

SSH and PAM then use that file during authentication. Flight MFA does
not enforce MFA by itself; enforcement is done through SSHd and PAM.

## Prerequisites

Install the Flight User Suite MFA package and its dependencies.

On Rocky 9, install:

- `google-authenticator`
- `qrencode`

On Ubuntu 24.04, install:

- `libpam-google-authenticator`
- `qrencode`

You also need:

- a working SSH service with PAM enabled
- user passwords configured for accounts that will use MFA
- SSH public key authentication configured for those accounts

## Configure Flight MFA

Review and update:

```text
/opt/flight/etc/mfa.config
```

The key settings are:

- `FLIGHT_MFA_SERVICE`: the service name shown in the authenticator app
- `FLIGHT_MFA_ISSUER`: the issuer shown in the authenticator app
- `FLIGHT_MFA_SECRETFILE`: optional override for the secret file location

Example:

```bash
FLIGHT_MFA_SERVICE=mycluster
FLIGHT_MFA_ISSUER=security@example.com
```

## Configure PAM

Add `pam_google_authenticator.so` to the SSH PAM stack in
`/etc/pam.d/sshd`.

The example shipped with Flight MFA expects a line equivalent to:

```text
auth required pam_google_authenticator.so secret=${HOME}/.config/flight/mfa.dat [authtok_prompt=Enter your TOTP: ] nullok
```

Key points:

- `secret=...` must match the file used by Flight MFA
- `nullok` allows users without an MFA secret file to continue logging in
  without a TOTP code
- `[authtok_prompt=Enter your TOTP: ]` customizes the SSH prompt the user
  sees

If you remove `nullok`, users without an enrolled token will be blocked at
login. That can be appropriate for strict environments, but it is usually
better to leave `nullok` in place during rollout so users can enroll
first.

The repository also includes an optional helper script:

```text
/opt/flight/usr/libexec/flight-mfa/mfa-warn.sh
```

This script prints a warning to users who have not yet generated their MFA
secret. It can be used as part of your PAM or session flow if you want to
encourage enrollment before making MFA mandatory.

To enable that warning during SSH sessions, add the following line to the
PAM configuration:

```text
session required pam_exec.so stdout /opt/flight/usr/libexec/flight-mfa/mfa-warn.sh
```

This causes PAM to run the helper during session setup and print the
warning banner when the user does not yet have an MFA secret file.

## Configure SSHd

Flight MFA is designed for SSH public key authentication combined with PAM.
Example configuration is shipped in:

```text
/opt/flight/usr/share/examples/flight-mfa/10-flight-mfa.conf
```

That example enables MFA for a specific client subnet and allows these
authentication combinations:

- `publickey,password` for users who do not yet have an MFA secret file
- `publickey,keyboard-interactive` for users who do have MFA configured,
  where keyboard-interactive is used to collect the password and TOTP

Important SSHd settings from the example are:

```text
UsePAM yes

Match Address 10.151.0.0/24
  PermitRootLogin no
  PasswordAuthentication yes
  KbdInteractiveAuthentication yes
  AuthenticationMethods publickey,keyboard-interactive publickey,password
```

These settings matter because:

- `UsePAM yes` is required; Flight MFA does not support `UsePAM no`
- `PasswordAuthentication yes` is needed when `publickey,password` is one
  of the allowed methods
- `KbdInteractiveAuthentication yes` is needed so PAM can prompt for the
  TOTP code
- `AuthenticationMethods` controls which combinations are accepted
- `Match Address` lets you enforce MFA only for selected source networks

Copy the example configuration to `/etc/ssh/sshd_config.d/` and edit the `Match
Address` line to include all of the subnets that should be configured to use
MFA.

After editing SSHd configuration, validate it and reload the service using
your site’s normal operational procedure.

## Recommended Rollout

A practical rollout is:

1. Install the package and dependencies.
2. Configure `/opt/flight/etc/mfa.config`.
3. Add the PAM line with `nullok`.
4. Enable the SSHd settings required for keyboard-interactive and PAM.
5. Ask users to run `flight mfa generate` and register their token.
6. Verify that enrolled users can log in with public key, password, and
   TOTP.
7. Optionally tighten policy later if you want to prevent logins for users
   who have not enrolled.

Using `Match Address` in SSHd is useful when you only want MFA from
untrusted networks while keeping simpler access from trusted internal
sources.

## What Changes for Users

Once MFA is configured and enforced:

- users still begin with SSH public key authentication
- users will also be prompted for their password
- enrolled users will be prompted for a TOTP code from their authenticator
  app
- users without an enrolled token may still be able to log in if PAM is
  configured with `nullok` and SSHd still allows `publickey,password`

If a user loses their authenticator device, they will need administrator
support to reset or replace their MFA secret before they can enroll a new
device.
