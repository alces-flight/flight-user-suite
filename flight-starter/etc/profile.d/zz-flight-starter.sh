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

if [ "${_nounset_set}" == "true" ]; then
  set -u
fi
unset _nounset_set
