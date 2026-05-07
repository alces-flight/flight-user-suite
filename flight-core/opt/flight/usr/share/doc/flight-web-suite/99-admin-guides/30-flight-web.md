---
admin: true
---
# Flight Web Suite

Flight Web Suite provides browser access to the Flight User Suite. Users can log
in using their cluster credentials and gain access to Flight User Suite tools
(where enabled) including Flight Desktop and howto guides.

## Prerequisites

Flight Web Suite requires:

* Python 3
* The [python-pam](https://pypi.org/project/python-pam/) library. Known as
  `python3-pampy` on Ubuntu. Known as `python3-pam` on Rocky 9 available from
  EPEL.

You can verify these prerequisites with:

```bash
flight web doctor
```

If the Flight desktop tool is enabled:

* `/usr/bin/websockify` to provide browser access to Flight Desktop sessions
* Optionally, `/usr/bin/import` from ImageMagick to enable screenshot capture
  for Flight Desktop sessions

You can verify these prerequisites with:

```bash
flight desktop doctor
```

## Configuration

Configuration is via a file located at `/opt/flight/etc/web-suite.yml`. The
default configuration file describes the available options.

User sessions are secured via a session secret, stored at
`/opt/flight/var/lib/web-suite/session-secret`. If this file is not present then
a random secret will be created automatically when Flight Web Suite starts for
the first time.  Changing the secret will invalidate all current sessions.

See below for details of configuring a reverse proxy, which is our recommended
deployment configuration.

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

* **Check dependencies:** as `root` run

  ```bash
  flight web doctor
  ```

## Accessing Flight Web

The output of `flight web start` will include the URL at which Flight Web Suite
can be accessed. Flight Web Suite uses the PAM `login` module to provide user
authentication; users can log in with their cluster username and password.

In order to access Flight tools via Web Suite, the CLI counterparts to those
tools will need to be enabled in the User Suite CLI tool (i.e. with
`flight tool enable`). You do not separately enable a tool in the CLI and Web
Suite.

## Reverse proxy

We recommend deployment behind a reverse proxy to provide TLS termination, and
not exposing the Flight Web Suite instance directly (especially if it is to be
made available over public Internet).

`nginx` is a suitable reverse proxy, for which an example configuration is
given below. You will need to ensure the values of `server_name`,
`ssl_certificate` and `ssl_certificate_key` are correct for your environment,
and that the value of `proxy_pass` (both occurrences) points to the correct
Flight Web Suite URL.

```text
# /etc/nginx/conf.d/web-suite.conf
server {
    listen 443 ssl;
    server_name login1.cluster.network;

    location / {
      proxy_pass http://localhost:8080;
      proxy_set_header X-Forwarded-For $remote_addr;
      proxy_set_header Host $host;
    }

    location /websockify {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;

        # Wait 24 hours without response before closing connection
        proxy_read_timeout 86400s;  # == 24 hours
        proxy_send_timeout 86400s;
    }
    ssl_certificate /etc/nginx/certs/localhost.crt;
    ssl_certificate_key /etc/nginx/certs/localhost.key;
}
```

Per nginx defaults, desktop sessions may disconnect if left idle within 60s of
inactivity. This will not harm the desktop session as it can be immediately
reconnected to. The `proxy_read_timeout` and `proxy_send_timeout` lines above
extend this idle period to 24 hours and can be adjusted to suit your environment
and requirements.
