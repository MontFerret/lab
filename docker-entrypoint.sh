#!/bin/sh

set -eu

is_lab_invocation() {
	if [ "$#" -eq 0 ]; then
		return 0
	fi

	case "$1" in
		run | version | help)
			return 0
			;;
		-*)
			return 0
			;;
	esac

	return 1
}

if is_lab_invocation "$@"; then
	./entrypoint.sh &
	exec ./lab "$@"
fi

exec "$@"
