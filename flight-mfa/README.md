# Flight MFA

The Flight MFA tool provides a facility for users to manage their
multi-factor authentication configuration.

## Prerequisites

You'll need to install some prerequisite packages.

### Rocky 9

 * Install the `google-authenticator` and `qrencode` packages from the EPEL repository.

### Ubuntu 24.04

 * Install the `libpam-google-authenticator` and `qrencode` packages.

## Usage

* Generate an MFA token
  ```sh
  flight mfa generate
  ```
* Show the generated token
  ```sh
  flight mfa show
  ```
