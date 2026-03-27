# Flight Starter

Flight Starter is responsible for activating the Flight environment in a shell.
It modifies the shell's `PATH` and `PS1` and sets a number of environment
variables.

After the Flight environment has been activated, by running `flight-start`, the
`flight` executable provided by `flight-core` is available on the shell's
`PATH`.  Further interaction with the Flight User Suite take place through that
`flight` executable.

When the Flight environment is deactivated, by running `flight-stop`,
modifications to the environment are cleaned up.


## Usage

Source the file `/etc/profile.d/zz-flight-starter.sh` and call the function
`flight-start` to activate the flight environment.  To deactivate the flight
environment call the function `flight-stop`.

### Development usage

During development the paths will be different.
