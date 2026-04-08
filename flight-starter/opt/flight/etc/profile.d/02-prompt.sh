#==============================================================================
# Copyright (C) 2019-present Concertim Ltd.
#
# This file is part of Flight User Suite.
#
# This program and the accompanying materials are made available under
# the terms of the Eclipse Public License 2.0 which is available at
# <https://www.eclipse.org/legal/epl-2.0>, or alternative license
# terms made available by Alces Flight Ltd - please direct inquiries
# about licensing to licensing@concertim.com.
#
# Flight Runway is distributed in the hope that it will be useful, but
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, EITHER EXPRESS OR
# IMPLIED INCLUDING, WITHOUT LIMITATION, ANY WARRANTIES OR CONDITIONS
# OF TITLE, NON-INFRINGEMENT, MERCHANTABILITY OR FITNESS FOR A
# PARTICULAR PURPOSE. See the Eclipse Public License 2.0 for more
# details.
#
# You should have received a copy of the Eclipse Public License 2.0
# along with Flight Runway. If not, see:
#
#  https://opensource.org/licenses/EPL-2.0
#
# For more information on Flight User Suite, please visit:
# https://github.com/alces-flight/flight-user-suite
#==============================================================================
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
