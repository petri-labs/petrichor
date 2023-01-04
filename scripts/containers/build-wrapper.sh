#!/usr/bin/env sh
set -x

export PATH=$PATH:/petrichord/petrichord
BINARY=/petrichord/petrichord
ID=${ID:-0}
LOG=${LOG:-petrichord.log}

if ! [ -f "${BINARY}" ]; then
	echo "The binary $(basename "${BINARY}") cannot be found."
	exit 1
fi

export PETRICHORDHOME="/petrichord/data/node${ID}/petrichord"

if [ -d "$(dirname "${PETRICHORDHOME}"/"${LOG}")" ]; then
  "${BINARY}" --home "${PETRICHORDHOME}" "$@" | tee "${PETRICHORDHOME}/${LOG}"
else
  "${BINARY}" --home "${PETRICHORDHOME}" "$@"
fi
