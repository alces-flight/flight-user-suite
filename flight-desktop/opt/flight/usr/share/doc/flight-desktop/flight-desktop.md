# What is Flight Desktop?

Flight Desktop, the tool that empowers you to create, manage, and access virtual
desktop sessions within your HPC cluster. This guide provides step-by-step
instructions for users to launch, monitor, and troubleshoot their remote desktop
environments.

Flight Desktop allows users to create and manage virtual desktop sessions of
various environment types. Key features include:

- **Isolation**: Each session has its own credentials, ensuring user privacy and
  security.
- **Accessibility**: Sessions are accessed remotely via VNC.
- **Collaboration**: If session details are shared, multiple users can connect
  to the same session simultaneously, enabling training and collaborative work.

## Desktop environment support

Flight Desktop comes with support for running sessions in the following
environments:

- xterm
- Gnome

Currently all sessions are backed by the X Window System.

## User Guide

### Prerequisites

You will need to be in an active Flight Environment for the following commands
to work. If your terminal session is not already in the environment then run:

```bash
flight-start
```

Refer to the Flight Environment documentation for further details.

### Managing Sessions

To view what sessions you have running:

```bash
flight desktop list
```

To start a new session of the specified type:

```bash
flight desktop start gnome --name mydesktop1
```

Replace `gnome` with the required environment type (as detailed above), and
`mydesktop1` with your desired session name.

To view connection details (IP, port, etc) for an existing session:

```bash
flight desktop show SESSION_NAME
```

Replace `SESSION_NAME` with the name of the session.

To kill a session:

```bash
flight desktop kill SESSION_NAME
```

Replace SESSION_NAME with the name of the session.

## How to connect to a session

While connecting to a session can vary depending on your system, network and
access configuration, the below provides some advice and possible methods to
connect to a remote desktop session.

### Direct

If the hostname/IP and port are directly accessible from your client then
entering the relevant details (host, port, username, password) into your VNC
client software should lead to a successful connection.

### SSH Tunnel

For tightly secured environments where VNC ports may not be directly available
it is possible to tunnel the VNC ports to your local system for connection.

1. Identify the session IP and port from `flight desktop show`
1. Use your preferred SSH client to tunnel to the VNC port:

   ```bash
   ssh -L 1234:localhost:5901 USERNAME@HPC_ENVIRONMENT_LOGIN
   ```

   - `1234` is the port on your local machine that you will be connecting to
   - `5901` is the port of your Flight Desktop session
   - `USERNAME` is the name you access the HPC Environment with
   - `HPC_ENVIRONMENT_LOGIN` is the hostname/IP you usually use to to connect to
     the HPC environment

1. Launch your VNC viewer application and connect to `localhost:5901` (or the
   port you used above).

The above example will work if the `HPC_ENVIRONMENT_LOGIN` system is where the
desktop session is running. Further tunnelling will be required if a different
system within the HPC environment is running the session. In which case, use the
"local machine" port (`1234` in the example above) for the first tunnel into
the HPC environment and use the VNC port on the subsequent login to the system
running the session.

## Troubleshooting

Below are a few steps that will help identify the cause of remote desktop
launch/access issues.

- Review the "System configuration recommendations" to ensure appropriate access
  is available to sessions.

- If you have access to them, check `/var/log/secure`, `/var/log/messages` and
  other system logs relating to your network and security configuration for any
  related errors.
