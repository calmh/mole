#!/bin/sh

# This should hopefully be enough for practical purposes
# and be small enough to be allowed in most installations.
ulimit -n 2048

exec mole.real $*
