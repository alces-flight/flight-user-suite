# Managed cluster

Contains a Flight User Suite hook to ease setup of managed clusters.  The
user-symlinks hook will create a number of symlinks to from the user's home
directory to other areas on the cluster.

An example is the symlink created from `$HOME/scratch` to `/scratch/$USER`
which is intended to be used for temporary data shared across cluster nodes.

## Usage

* Enable the hook: `/opt/flight/bin/flight hooks enable login user-symlinks`.
* Activate the flight environment: `flight-start`.
