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

* Edit config file directly.
  * Where is it?
  * What do the different values mean?
  * What are the permitted values for each setting?

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
