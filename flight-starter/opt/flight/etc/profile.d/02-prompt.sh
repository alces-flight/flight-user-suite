export FLIGHT_ORIG_ENV_PS1="${PS1}"

FLIGHT_BLUE="38;2;32;159;206"

# Start with the basics - we'll end up injecting a space between \h and \W later
PS1="[\u@\h\W]\\$ "

# Prepend an alces blue <flight>
PS1="\[\033[${FLIGHT_BLUE}m\]\$(__flight_ps1_active \"<%s>\")\[\033[00m\] ${PS1}"

FLIGHT_PS1="$(
    "${FLIGHT_ROOT}"/libexec/flight-starter/augment-bash-prompt \
        "$PS1" \
        '$(__flight_ps1_clustername "(%s) ")' \
        "$FLIGHT_BLUE" \
        2>/dev/null
    )"
if [ $? -eq 0 ] ; then
    PS1="${FLIGHT_PS1}"
fi

__flight_ps1_clustername() {
    local printf_format='(%s)'
    case "$#" in
        0|1)	printf_format="${1:-$printf_format}"
            ;;
        *)	return 0
            ;;
    esac

    source "${FLIGHT_ROOT}"/etc/flight-starter.config
    local cluster_name flight_string
    cluster_name="${FLIGHT_STARTER_CLUSTER_NAME:-your cluster}"
    if [ "${cluster_name}" != "your cluster" ] ; then
        flight_string="${cluster_name}"
    fi

    if [ "${flight_string}" != "" ]; then
        printf -- "$printf_format" "$flight_string"
    fi
    unset $(declare | grep ^FLIGHT_STARTER | cut -f1 -d= | xargs)
}

__flight_ps1_active() {
    local printf_format='(%s)'
    case "$#" in
        0|1)	printf_format="${1:-$printf_format}"
            ;;
        *)	return 0
            ;;
    esac

    local flight_string
    if [ "${FLIGHT_ACTIVE}" == "true" ] ; then
        flight_string="flight"
    fi

    if [ "${flight_string}" != "" ]; then
        printf -- "$printf_format" "$flight_string"
    fi
}

FLIGHT_DEFINED_SYMBOLS+=(__flight_ps1_active __flight_ps1_clustername)
unset FLIGHT_PS1 FLIGHT_BLUE
