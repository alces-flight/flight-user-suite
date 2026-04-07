# Managed cluster documentation

## User symlinks

When you first activate the Flight environment a number of symlinks will be
created from your home directory to other locations on your HPC environment.
These symlinks are designed to ensure optimal use of your HPC environment.

* **Scratch**: A link is created from `~/scratch` to `/scratch/<user>`. This
  directory is shared across the cluster and should be used for any temporary
  data files created by your HPC jobs.
