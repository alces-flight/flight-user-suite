---
admin: true
---
# What is Flight Web?

* Provides browser access to the Flight User Suite, including desktop and howto guides.

## Configuring Flight Web

TBC.

## Usage

* **Enable:** as `root` run
  ```bash
  flight tools enable --admin-only web
  ```
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
