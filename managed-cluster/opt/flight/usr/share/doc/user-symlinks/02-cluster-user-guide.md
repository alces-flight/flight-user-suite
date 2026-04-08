# HPC Environment User Guide

Welcome to your Managed Cluster. This guide contains tips, tricks and insight
for your HPC Environment.

## Storage Information

Your HPC Environment has multiple storage areas of different capacities and
functions. 

In order to best manage your data it is worth understanding and following the
usage guidelines for these: 

- `/data`: This storage areas provides long-term storage for data. It is
  recommended to use this space for results or any other data that you wish to
  keep after running your workload.
- `/scratch`: This storage area is a large storage space intended for temporary
  storage of workload information before and during execution. It is not
  intended for long-term data storage. 
- `/users`: This storage area provides private storage for each user.

## Data Management

When it comes to managing your workload data it is worth familiarising yourself
with the Storage Information outlined above. 

Your private user storage area is your home directory. In this directory you
will find a symbolic link named `scratch` which will take you to your
user-specific area on the `/scratch` storage area.

Your private user storage areas has usage quotas on it. This is to promote
fair-share of the storage space amongst all users.

### Going Over Quota

It is possible to be over quota in multiple ways. There are limits to the amount
of disk space used (“Space Utilisation”) and number of files created (“File
Utilisation”).

You have a SOFT limit and a HARD limit. The HARD limit is typically set at 1.5x 
the SOFT limit. 

1. When you first exceed a quota you will enter a grace period. You will be able
   continue to create files and use space as normal until either the grace
   period expires or you reach your HARD limit. 
2. If your grace period expires, you will hit your SOFT limit and will no longer
   be able to create new files or use more space. This may effect your workloads
   and HPC service experience.
3. If at any point you exceed your HARD limit, you will no longer be able to
   create new files or use more space even during any initial grace period.

If you are having issues with your quota please contact your HPC Environment
team..

## Running workloads

It is important to consider where and how you are running workloads on the HPC
Environment.

- **Login Nodes**: You will likely be accessing the HPC Environment via a login
  node. This can be used for creating interactive desktop sessions for
  basic/light usage. Due to this being a shared point of access for users there
  are policies in place to manage the load by preventing processes from
  overutilising resources. 
- **Scheduler**: The scheduler provides resource management for the HPC
  Environment. Submitting jobs to the scheduler queue will allow you to reserve
  and utilise resources for your workloads whether they are interactive, batch
  or arrays jobs.

## Getting Help

If you need help using the HPC Environment please reach out to your IT Team.
