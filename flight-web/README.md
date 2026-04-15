# Flight Web

Flight Web provides browser based access to the Flight User Suite tools.

## Usage

* Enable the Flight Web management tool. The `--admin-only` flag means that only the root user will be able to manage the web service.
  ```sh
  sudo flight tools enable --admin-only web
  ```
* Start Flight Web
  ```sh
  sudo flight web start
  ```
* Stop Flight Web
  ```sh
  sudo flight web stop
  ```

When started, Flight Web will print the URL at which it can be accessed. Point
your browser at that URL to use it.
