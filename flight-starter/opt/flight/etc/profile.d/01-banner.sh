#==============================================================================
# Copyright (C) 2026 Stephen F Norledge & Alces Software Ltd & Concertim Ltd
#
# This file is part of Flight User Suite.
#
# This program and the accompanying materials are made available under
# the terms of the Eclipse Public License 2.0 which is available at
# <https://www.eclipse.org/legal/epl-2.0>, or alternative license
# terms made available by Alces Software Ltd - please direct inquiries
# about licensing to licensing@alces-software.com.
#
# Flight User Suite is distributed in the hope that it will be useful, but
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, EITHER EXPRESS OR
# IMPLIED INCLUDING, WITHOUT LIMITATION, ANY WARRANTIES OR CONDITIONS
# OF TITLE, NON-INFRINGEMENT, MERCHANTABILITY OR FITNESS FOR A
# PARTICULAR PURPOSE. See the Eclipse Public License 2.0 for more
# details.
#
# You should have received a copy of the Eclipse Public License 2.0
# along with Flight User Suite. If not, see:
#
#  https://opensource.org/licenses/EPL-2.0
#
# For more information on Flight User Suite, please visit:
# https://github.com/alces-flight/flight-user-suite
#==============================================================================
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
