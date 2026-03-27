# Flight Howto

Flight Howto is a tool to list and view Markdown documents that form the basis
of Flight User Suite's online documentation.

## Usage

* List available documentation

```sh
flight howto list
```

* View documentation

```sh
flight howto show <document>
```

`<document>` should be a filename as listed by `flight howto list`.

## Providing documentation

All Markdown files (`*.md`) in the `${FLIGHT_ROOT}/usr/share/doc/howtos-enabled`
directory or subdirectories thereof are included in the listing.

When a tool is enabled by `flight tools`, it will automatically symlink
Markdown documentation from the `${FLIGHT_ROOT}/usr/share/doc/<tool>` directory
into `howtos-enabled` to make it available via `flight howto`, but documentation
can also be created directly in that directory (or elsewhere, and manually
symlinked in).

## Rendering

Markdown rendering is handled by [Glamour](https://github.com/charmbracelet/glamour)
and uses the stock `dark` theme.
