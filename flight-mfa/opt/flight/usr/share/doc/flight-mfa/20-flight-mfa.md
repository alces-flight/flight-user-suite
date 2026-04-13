---
name: "Flight MFA"
---
# Flight MFA

Flight MFA lets you set up a time-based one-time password (TOTP) for your
account. A TOTP code is the short changing number shown in an
authenticator application on your phone or other device.

When your administrator enables MFA for SSH access, logging in may require
more than your SSH key or password alone. You may be asked for:

- your SSH key
- your account password
- a TOTP code from your authenticator application

This means a stolen password or SSH key is not enough on its own to access
your account.

## Before You Start

You need:

- access to a TOTP authenticator application
- a terminal session on the system with Flight User Suite available

Examples of TOTP applications include Google Authenticator and other apps
that can scan a QR code or accept a manual setup key.

## Generate Your MFA Secret

Run:

```bash
flight-start
flight mfa generate
```

This generates your MFA secret and displays it as:

- a QR code in the terminal, if the terminal is wide enough
- a manual setup key, which you can type into your authenticator app if
  needed

Scan the QR code with your authenticator app, or enter the key manually.
Once it has been added, your app will begin generating 6-digit TOTP codes
for this system.

## Show an Existing MFA Secret

If you need to register the same token on another device, run:

```bash
flight-start
flight mfa show
```

This displays the existing QR code or manual setup key again. It does not
generate a new secret.

## Regenerate Your MFA Secret

If you need to replace the old token with a new one, run:

```bash
flight-start
flight mfa generate --force
```

This replaces the existing secret. After doing this, the old token in your
authenticator app will stop working and you must register the new one.

## What Changes When MFA Is Enabled

The exact login flow depends on how your administrator has configured SSH
and PAM, but a common setup is:

1. SSH verifies your public key.
2. You are prompted for your account password.
3. You are prompted for a TOTP code.

If MFA has been enabled by your administrator but you have not yet set up
your token, one of two things usually happens:

- you may still be able to log in with your SSH key and password while MFA
  enrollment is being rolled out
- you may be blocked from login until you configure MFA, if your site has
  made MFA mandatory

## Troubleshooting

If `flight mfa generate` or `flight mfa show` does not work as expected:

- make sure `flight-start` has been run if your environment requires it
- make sure your terminal is wide enough to display the QR code
- use the printed setup key if the QR code cannot be shown
- check that your system clock and your authenticator device clock are
  correct

If your TOTP code is rejected repeatedly, contact your administrator. The
stored secret may need to be reset and generated again.
