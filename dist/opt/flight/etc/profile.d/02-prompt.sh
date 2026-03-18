export FLIGHT_ORIG_ENV_PS1="${PS1}"

if [ "$PS1" = "\\s-\\v\\\$ " ]; then
  # prompt hasn't been set yet, give it a default
  PS1="[\u@\h\$(__flight_ps1) \W]\\$ "
fi

__flight_ps1() {
    local exit=$?
    local printf_format=' (%s)'
    case "$#" in
        0|1)	printf_format="${1:-$printf_format}"
            ;;
        *)	return $exit
            ;;
    esac

    local cluster_name
    cluster_name="your cluster"
    local flight_string="${cluster_name}"

    printf -- "$printf_format" "$flight_string"
}
