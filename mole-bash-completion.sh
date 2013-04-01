#!/bin/bash

# Find the completion file.

NPMPREF=$(npm config get prefix)
COMP="$NPMPREF"/lib/node_modules/mole/bash_completion.d/mole
if [[ ! -f $COMP ]] ; then
        echo $COMP is missing.
        echo I\'m not sure why, but it\'s a bad sign. Is mole installed globally\?
        exit -1
fi

# Verify that we're not root

if [[ $(id -u) == 0 ]] ; then
        echo This script want\'s to modify your \~/.bash_profile. So don\'t run it as root.
        exit -1
fi

# Check for existing installation

if ( grep -q "$COMP" ~/.bash_profile ) ; then
        echo It seems it\'s already installed.
        exit 0
else
        echo "# Automatically added by mole-bash-completion:" >> ~/.bash_profile
        echo "source $COMP" >> ~/.bash_profile
        echo Installed. Double check your \~/.bash_profile if you like, and restart your shell.
fi
