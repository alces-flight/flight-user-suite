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
