#!/bin/sh

# Find the completion file.

NPMPREF=$(npm config get prefix)
COMP=$NPMPREF/lib/node_modules/mole/mole.plugin.zsh
if [ ! -f $COMP ] ; then
        echo $COMP is missing.
        echo I\'m not sure why, but it\'s a bad sign. Is mole installed globally\?
        exit -1
fi

# Check that we're running zsh.

case "$SHELL" in
*zsh)
        # It's all good
        ;;
*)
        echo You seem to be running $SHELL, not zsh.
        echo That\'s a valid choice, but the completion functions for mole won\'t work.
        exit -1
        ;;
esac

# Verify that we're not root

if [ $(id -u) == "0" ] ; then
        echo This script want\'s to modify your \~/.zshrc. So don\'t run it as root.
        exit -1
fi

# Check for existing installation

if ( grep -q $COMP ~/.zshrc ) ; then
        echo It seems it\'s already installed.
        exit 0
else
        echo "# Automatically added by mole-zsh-completion:" >> ~/.zshrc
        echo "source $COMP" >> ~/.zshrc
        echo Installed. Double check your \~/.zshrc if you like, and restart your shell.
fi
