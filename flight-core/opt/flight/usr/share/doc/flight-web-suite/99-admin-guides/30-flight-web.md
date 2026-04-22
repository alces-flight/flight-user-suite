---
admin: true
---
# What is Flight Web?

* Provides browser access to the Flight User Suite, including desktop and howto guides.

## Setup

Requires:

* Python3
* The [python-pam](https://pypi.org/project/python-pam/) library. Known as `python3-pampy` on Ubuntu. Known as `python3-pam` on Rocky 9 available from EPEL.

## Configuration

* Edit config file at `/opt/flight/etc/web-suite.yml` directly.
  * What do the different values mean?
  * What are the permitted values for each setting?

The session secret is stored in at
`/opt/flight/var/lib/web-suite/session-secret`. It will be created
automatically when web suite starts for the first time.  Changing it will
invalidate all sessions.

## Usage

* **Start:** as `root` run
  ```bash
  flight web start
  ```
* **Stop:** as `root` run
  ```bash
  flight web stop
  ```
* **Get status:** as `root` run
  ```bash
  flight web status
  ```

## Accessing Flight Web

* Port and interface are configured in "Configuring Flight Web"
* URL is printed in output of `flight web start`.  It will be the same each time unless reconfigured.
* URL can be obtained from `flight web status`.

## User authentication

* Uses PAM `login` module.
* Users provide cluster username and password.
