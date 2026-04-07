# SSH keypair generation

SSH keypair generation is a hook that generates a passless SSH key that can be
used to SSH around the cluster.  Ensuring that the key is authorised on each
node is outside the scope of this hook; we assume that adding it to
`~/.ssh/authorized_keys` is sufficient.

## Usage

* Enable the hook: `/opt/flight/bin/flight hooks enable login ssh-keypair-generation`.
* Activate the flight environment: `flight-start`.
