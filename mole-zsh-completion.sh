#!/bin/sh

# Find the completion file.

NPMPREF=$(npm config get prefix)
COMP=$NPMPREF/lib/node_modules/mole/mole.zsh-completion
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

# It might seem like we'd want local/share before share, but that's not the case.

for DIR in /usr/share/zsh/site-functions /usr/local/share/zsh/site-functions ; do
        if [ -d $DIR ] ; then
                DEST=$DIR/_mole

                # We're going to need to be root

                if [ $(id -u) != "0" ] ; then
                        echo This script is going to create the file $DEST.
                        echo To do that, we need to be root. Please run this script as:
                        echo
                        echo '   ' sudo mole-zsh-completion
                        echo
                        exit -1
                fi

                SCR=$(mktemp /tmp/_mole.XXXXXX)
                echo source $COMP > $SCR

                cp $SCR $DEST
                chmod 755 $DEST

                echo Installed. Please make sure you have the following somewhere in your \~/.zshrc:
                echo
                echo '   ' autoload -U compinit
                echo '   ' compinit
                echo
                echo Then, restart zsh and enjoy the tabbing.
                exit 0
        fi
done
