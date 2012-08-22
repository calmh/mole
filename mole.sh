#!/bin/sh

MAX=4096

# Try to find the largest usable file descriptor limit or $MAX, whichever is
# smaller.

HARD=$(ulimit -Hn)

if [ "$HARD" = "unlimited" ] ; then
        HARD=$MAX
elif [ "$HARD" -gt "$MAX" ] ; then
        HARD=$MAX
fi

ulimit -n $HARD
exec mole.real $*
