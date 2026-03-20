(
    _shell_is_interactive=false
    if [ "${-#*i}" != "$-" ]; then
        _shell_is_interactive=true
    fi

    if [ "$_shell_is_interactive" == true -o "$1" == "--force" ] ; then
        source "${FLIGHT_ROOT}"/etc/flight-starter.config
        if [ -f /etc/redhat-release ]; then
            release="$(cut -f1,2,4 -d' ' /etc/redhat-release)"
        elif [ -f /etc/lsb-release ]; then
            . /etc/lsb-release
            release="${DISTRIB_DESCRIPTION:-${DISTRIB_ID} ${DISTRIB_RELEASE}}"
        fi
        ${FLIGHT_ROOT}/libexec/flight-starter/banner \
            "${FLIGHT_STARTER_CLUSTER_NAME:-your cluster}" \
            "${FLIGHT_STARTER_PRODUCT} ${FLIGHT_STARTER_RELEASE}" \
            "${release}"
    fi
)
