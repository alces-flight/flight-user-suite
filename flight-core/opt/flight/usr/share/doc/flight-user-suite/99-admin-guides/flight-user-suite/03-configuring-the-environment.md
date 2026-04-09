---
admin: true
---
# Configuring Flight User Suite

## Flight User Suite

Using `flight config` you can set the default behaviour of the Flight
environment for end-users. 

Currently supported configuration options:
- `autostart`: This controls whether the Flight User Suite environment is
  activated by default on login. Valid options are `on` and `off`.

To set a configuration option:
```
flight config set --global OPTION VALUE
```

_Note: Users can set these options for themselves. For example, if the global
'autostart' option is 'on' but the user sets it to 'off' for themselves then 
the Flight environment will not be active for them by default when they login_

## Prompt

To configure the prompt shown upon activation of the Flight environment, edit 
`/opt/flight/etc/flight-starter.config`. This will allow setting of things like 
the cluster name, environment description and release information. 
