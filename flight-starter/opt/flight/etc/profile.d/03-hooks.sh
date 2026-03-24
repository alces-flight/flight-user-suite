# Run any enabled activation hooks when the Flight environment is activated.
if [ -d "${FLIGHT_ROOT}"/usr/lib/hooks/activation ]; then
  shopt -s nullglob
  for hook in "${FLIGHT_ROOT}"/usr/lib/hooks/activation/*; do
      if [ -x "${hook}" ] ; then
          "${hook}"
      fi
  done
  shopt -u nullglob
fi
unset hook
