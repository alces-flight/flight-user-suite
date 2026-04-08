# Managed Cluster Documentation

## User symlinks

When you first activate the Flight environment, the system automatically
creates symbolic links (symlinks) in your home directory. These links are
designed to help you navigate and use the different storage areas available on
your HPC environment.

* **Scratch Storage** (`~/scratch`): Points to `/scratch/<username>`.
  * **Purpose**: Designed for high-performance storage of large datasets and temporary job files.
  * **Scope**: This directory is shared across the cluster, ensuring your jobs can access data from any node.

## Administration & Setup

The target directories (e.g., `/scratch/<username>`) must exist for the symlinks to function. You can create these manually or automate the process using Flight User Suite.

### Automated Directory Creation

To automate directory creation upon login, configure PAM to execute the setup
script when a new session starts, by adding the following line to
`/etc/pam.d/sshd`:

```
# Create user-specific directories for home-directory symlinks
session optional pam_exec.so seteuid /opt/flight/usr/libexec/managed-cluster/create-user-symlink-target-dirs
```

### Configuration

To define which symlinks are created or to add new ones, edit the following configuration file:

* `/opt/flight/etc/user-symlinks.config`
