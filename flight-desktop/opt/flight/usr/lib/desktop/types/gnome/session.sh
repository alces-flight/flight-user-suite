xdg_config_home() {
    echo "${XDG_CONFIG_HOME:-$HOME/.config}"
}

kill_on_script_exit=
if [ "$1" == "--kill-on-script-exit" ] ; then
    kill_on_script_exit=true
    shift
fi

# 'Xterm*vt100.pointerMode: 0' is to ensure that the pointer does not
# disappear when a user types into the xterm.  In this situation, some
# VNC clients experience a 'freeze' due to a bug with handling
# invisible mouse pointers (e.g. MacOS Screen Sharing).
echo 'XTerm*vt100.pointerMode: 0' | xrdb -merge
vncconfig -nowin &
# Disable gnome-screensaver
gconftool-2 --set -t boolean /apps/gnome-screensaver/idle_activation_enabled false

# Create flag file to skip initial setup
mark_initial_setup_done() {
  local setup_file
  setup_file="$(xdg_config_home)/gnome-initial-setup-done"
  if [ ! -f "${setup_file}" ]; then
    echo -n "yes" > "${setup_file}"
  fi
}

mark_initial_setup_done
if [ -f /etc/redhat-release ] && ! grep -q 'release 9' /etc/redhat-release; then
  export GNOME_SHELL_SESSION_MODE=classic
  _GNOME_PARAMS="--session=gnome-classic"
fi

if [ "$1" ]; then
  gnome-terminal --wait -- "$@" &
  gnome_terminal_pid=$!
fi

gnome-session ${_GNOME_PARAMS} &
gnome_session_pid=$!

if [ "$kill_on_script_exit" == true ] ; then
    wait $gnome_terminal_pid
    sleep 2
else
    wait $gnome_session_pid
fi
