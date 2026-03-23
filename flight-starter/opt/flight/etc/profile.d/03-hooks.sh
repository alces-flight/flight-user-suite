# Run any enabled hooks when the Flight environment is activated.
if [ -d "${FLIGHT_ROOT}"/usr/lib/hooks ]; then
  shopt -s nullglob
  for hook in "${FLIGHT_ROOT}"/usr/lib/hooks/*; do
      if [ -x "${hook}" ] ; then
          "${hook}"
      fi
  done
  shopt -u nullglob
fi
unset hook
