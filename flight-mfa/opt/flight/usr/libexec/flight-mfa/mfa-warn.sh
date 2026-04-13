#!/bin/bash
#==============================================================================
# Copyright (C) 2026-present Alces Flight Ltd.
#
# This file is part of Flight User Suite.
#
# This program and the accompanying materials are made available under
# the terms of the Eclipse Public License 2.0 which is available at
# <https://www.eclipse.org/legal/epl-2.0>, or alternative license
# terms made available by Alces Flight Ltd - please direct inquiries
# about licensing to licensing@alces-flight.com.
#
# Flight MFA is distributed in the hope that it will be useful, but
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, EITHER EXPRESS OR
# IMPLIED INCLUDING, WITHOUT LIMITATION, ANY WARRANTIES OR CONDITIONS
# OF TITLE, NON-INFRINGEMENT, MERCHANTABILITY OR FITNESS FOR A
# PARTICULAR PURPOSE. See the Eclipse Public License 2.0 for more
# details.
#
# You should have received a copy of the Eclipse Public License 2.0
# along with Flight MFA. If not, see:
#
#  https://opensource.org/licenses/EPL-2.0
#
# For more information on Flight MFA, please visit:
# https://github.com/alces-flight/flight-user-suite
#==============================================================================

FLIGHT_ROOT=${FLIGHT_ROOT:-/opt/flight}
if [ -f "${FLIGHT_ROOT}"/etc/mfa.config ]; then
    . "${FLIGHT_ROOT}"/etc/mfa.config
else
    echo "${FLIGHT_PROGRAM_NAME:-mfa}: could not find configuration file"
    exit 1
fi

eval HOME=~$PAM_USER
_flightmfa_SECRETFILE="${HOME}/${FLIGHT_MFA_SECRETFILE:-.config/flight/mfa.dat}"

if [[ ! -f ${_flightmfa_SECRETFILE} ]] ; then
    cat << 'EOF'
============= MFA NOT CONFIGURED ==============

No MFA configuration found.

Please run following to configure MFA:

    flight mfa generate

===============================================
EOF
fi
