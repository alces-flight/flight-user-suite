# Flight Core

Flight Core is responsible for managing access to the Flight User Suite (FUS).

It contains builtin root-only commands to list, enable and disable FUS tools
and hooks. It also serves as the entry point for running FUS tools for both
root and non-root users.

## Usage

* List all available tools
  ```sh
  flight tools list
  ```
* List all enabled tools
  ```sh
  flight tools list --enabled
  ```
* Enable a tool
  ```sh
  flight tools enable <tool>
  ```
* Disable a tool
  ```sh
  flight tools disable <tool>
  ```
* Run a tool
  ```sh
  flight <tool> [tool arguments]
  ```
* List all available hooks
  ```sh
  flight hooks list
  ```
* List all enabled hooks
  ```sh
  flight hooks list --enabled
  ```
* Enable a hook
  ```sh
  flight hooks enable <event> <hook>
  ```
* Disable a hook
  ```sh
  flight hooks disable <event> <hook>
  ```
