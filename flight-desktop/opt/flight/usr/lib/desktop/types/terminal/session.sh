# 'Xterm*vt100.pointerMode: 0' is to ensure that the pointer does not
# disappear when a user types into the xterm.  In this situation, some
# VNC clients experience a 'freeze' due to a bug with handling
# invisible mouse pointers (e.g. OSX Screen Sharing).
echo 'XTerm*vt100.pointerMode: 0' | xrdb -merge
vncconfig -nowin &
xsetroot -fg '#2794d8' -bitmap <(
cat <<EOF
#define root_weave_width 4
#define root_weave_height 4
static char root_weave_bits[] = {
   0x07, 0x0d, 0x0b, 0x0e};
EOF
)

if [ "$1" == "--kill-on-script-exit" ] ; then
    shift
fi
if [ -n "$1" ]; then
  xterm -e "$@"
else
  xterm
fi
