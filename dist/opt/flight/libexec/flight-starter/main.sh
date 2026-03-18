if [ "$1" == "start" ]; then
    FLIGHT_DEFINED_SYMBOLS=(FLIGHT_ACTIVE)
    FLIGHT_ON_EXIT=()
    FLIGHT_ADDED_PATHS=()
    export FLIGHT_DEFINED_SYMBOLS FLIGHT_ADDED_PATHS FLIGHT_ON_EXIT

    _shell_is_interactive=1
    if [ "${-#*i}" != "$-" ]; then
        _shell_is_interactive=0
    fi

    # Source profile scripts.
    if [ -d ${FLIGHT_ROOT}/etc/profile.d ]; then
        if [ "${-#*e}" != "$-" ]; then
            _errexit_set=true
        fi
        shopt -s nullglob
        for i in ${FLIGHT_ROOT}/etc/profile.d/*.sh ; do
            if [ -r "$i" ]; then
                # Ensure flight profile can't cause errexit
                set +e
                if $_shell_is_interactive ; then
                    . "$i"
                else
                    . "$i" >/dev/null 2>&1
                fi
            fi
        done
        shopt -u nullglob
        if [ "${_errexit_set}" == "true" ]; then
            set -e
        fi
        unset i _set_errexit
    fi

    # Add FLIGHT_ROOT/bin to PATH
    if [[ ":$PATH:" != *":${FLIGHT_ROOT}/bin:"* ]]; then
        PATH="${PATH}:${FLIGHT_ROOT}/bin"
        FLIGHT_ADDED_PATHS+=(${FLIGHT_ROOT}/bin)
    fi

    export FLIGHT_ACTIVE=true

    if [ "$(type -t flight-stop)" != "function" ]; then
        flight-stop() {
            source "${FLIGHT_ROOT}"/libexec/flight-starter/main.sh stop
        }
        export -f flight-stop
    fi
    unset -f flight-start

    if $_shell_is_interactive ; then
        echo "Flight environment is now active." 1>&2
    fi
    unset _shell_is_interactive

elif [ "$1" == "stop" ]; then
    if [ "${-#*i}" != "$-" ]; then
        _shell_is_interactive=0
    else
        _shell_is_interactive=1
    fi

    # Call any registered exit hooks.
    for a in "${FLIGHT_ON_EXIT[@]}"; do
        $a
    done
    unset FLIGHT_ON_EXIT a

    # Reset environment variables to original value.
    for a in $(env | grep '^FLIGHT_ORIG_ENV_' | cut -f1 -d= | xargs); do
        tgt=$(echo "$a" | cut -f4- -d'_')
        eval "$tgt=\"${!a//\\/\\\\}\""
        unset tgt
        unset $a
    done

    for a in "${FLIGHT_DEFINED_SYMBOLS[@]}"; do
        unset $a
    done
    unset FLIGHT_DEFINED_SYMBOLS a

    # Remove any paths that have been added to PATH.
    if [ ${#FLIGHT_ADDED_PATHS[@]} -gt 0 ] ; then
        for a in "${FLIGHT_ADDED_PATHS[@]}"; do
            new_path=""
            IFS=: read -a paths <<< "${PATH}"
            for p in "${paths[@]}"; do
                if [ "${p}" != "${a}" ] ; then
                    new_path="${new_path}:$p"
                fi
            done
            PATH="${new_path}"
        done
    fi
    unset FLIGHT_ADDED_PATHS a new_path

    if [ "$(type -t flight-start)" != "function" ]; then
        flight-start() {
            source "${FLIGHT_ROOT}"/libexec/flight-starter/main.sh start
        }
        export -f flight-start
    fi
    unset -f flight-stop

    if $_shell_is_interactive; then
        echo "Flight environment is now inactive." 1>&2
    fi
    unset _shell_is_interactive
fi
