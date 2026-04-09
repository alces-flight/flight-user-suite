---
admin: true
---
## System configuration recommendations

Flight desktop is ideal for running in closed HPC environments. This ensures
that access security is strictly controlled by network and authentication
policies.

Before using Flight Desktop, verify that your system meets the following
criteria:

- SELinux is disabled or appropriately configured to allow remote connections
- The system, network and platform firewalls are configured to allow in a range
  of VNC ports (a range starting at `5901`)

You will also need to ensure that certain dependencies are installed on your
system. Flight desktop can provide a report highlighting any missing
dependencies, by running:

```bash
flight desktop doctor
```
