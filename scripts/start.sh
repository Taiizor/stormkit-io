#!/usr/bin/env bash
set -e

DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
WORKDIR="$(dirname "${DIR}")"

cd "${WORKDIR}"

if ! command -v overmind &>/dev/null; then
    echo "overmind could not be found"
    exit
fi

if ! command -v tmux &>/dev/null; then
    echo "tmux could not be found"
    exit
fi

echo "Loading environment variables from from .env file".

if [ -f "$WORKDIR/.env" ]; then
  export $(cat .env | xargs)
fi

export STORMKIT_PROJECT_ROOT=$(pwd)
export GO111MODULE=on
export CGO_ENABLED=1

go mod download

go build -o $WORKDIR/bin/runner $WORKDIR/src/ee/runner

overmind s -f Procfile

