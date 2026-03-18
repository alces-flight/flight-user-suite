The following environment variables are defined by the flight environment.

* `FLIGHT_ACTIVE` : set to true if the flight environment is active in the current shell.
* `FLIGHT_DEFINED_SYMBOLS` : array of symbols that have been defined by the flight environment. When the environment is exited, they are all undefined.
* `FLIGHT_ADDED_PATHS` : array of paths that the flight environment added to
PATH. They are removed from PATH when the environment exits.
* `FLIGHT_ORIG_ENV_*` : family of environment variables each containing the
original environment variable set whent the flight environment was activated.
E.g., `FLIGHT_ORIG_ENV_PS1` contains the original setting for `PS1`.  Any
variables registered here will be reset when the flight environment exits.
* `FLIGHT_ON_EXIT` : array of functions to evaluate when the flight environment
is exited. The intended use is to allow different parts of the system to define
their own clean up routines.
