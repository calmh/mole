_mole () {
	typeset -A opt_args
	typeset -A val_args
	local context state line
	local tunneldir ret
	tunneldir=~/.mole/tunnels

	if [[ $#words == 2 ]] ; then
		local -a commands
		commands=(
			'dig:dig a tunnel'
			'list:list tunnels'
			'pull:pull tunnel defs'
			'push:push a tunnel def'
			'export:export a tunnel def'
			'gettoken:generate a new token'
			'install:install an optional package'
			'register:register with a server'
			'deluser:delete a user {admin}'
			'newuser:create a user {admin}'
			'version:check for updates'
			'server:start a server instance'
		)
		_describe -t commands command commands && return 0
	elif [[ $#words -ge 3 ]] ; then
		local -a common
		common=( '-d[debug]' )
		case $words[2] in
			(dig)
			_arguments \
				':command' \
				':file:_files -W $tunneldir -g \*.ini\(:r\)' \
				'-l+[local file]:file:_files -g \*.ini' \
				$common \
				&& return 0
			;;
			(push)
			_arguments \
				':command' \
				':file:_files -g \*.ini\(:r\)' \
				$common \
				&& return 0
			;;
			(export)
			_arguments \
				':command' \
				':tunnel:_files -W $tunneldir -g \*.ini\(:r\)' \
				':file:_files -g \*.ini' \
				$common \
				&& return 0
			;;
			(*)
			_arguments \
				':command' \
				$common \
				&& return 0
			;;
		esac
	fi
}

compdef _mole mole

