# Flight Core

Flight Core is responsible for managing access to the Flight User Suite (FUS).

It contains builtin commands to list, enable and disable FUSuite tools.  It
also serves as the entry point for running FUS tools.

## Usage

* List all available tools
  ```sh
  flight list
  ```
* List all enabled tools
  ```sh
  flight list --enabled
  ```
* Enable a tool
  ```sh
  flight enable <tool>
  ```
* Disable a tool
  ```sh
  flight disable <tool>
  ```
* Run a tool
  ```sh
  flight <tool> [tool arguments]
  ```
