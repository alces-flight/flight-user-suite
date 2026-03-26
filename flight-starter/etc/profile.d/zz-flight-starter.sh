if [ -z "$BASH_VERSION" ]; then
  return
fi

export FLIGHT_ROOT=${FLIGHT_ROOT:-/opt/flight}

if [ "${-#*u}" != "$-" ]; then
  set +u
  _nounset_set=true
fi
IFS=: read -a xdg_config <<< "${XDG_CONFIG_HOME:-$HOME/.config}:${XDG_CONFIG_DIRS:-/etc/xdg}"
for a in "${xdg_config[@]}"; do
  if [ -e "${a}"/flight.config ]; then
    source "${a}"/flight.config
    break
  fi
done

if [ "$(type -t flight-start)" != "function" ]; then
  flight-start() {
    source "${FLIGHT_ROOT}"/libexec/flight-starter/main.sh start
  }
  export -f flight-start
fi

# Run any enabled login hooks when a login shell is created.
if [ -d "${FLIGHT_ROOT}"/usr/lib/hooks/login ]; then
  shopt -s nullglob
  for hook in "${FLIGHT_ROOT}"/usr/lib/hooks/login/*; do
      if [ -x "${hook}" ] ; then
          "${hook}"
      fi
  done
  shopt -u nullglob
fi
unset hook

if [ "${_nounset_set}" == "true" ]; then
  set -u
fi
unset _nounset_set
