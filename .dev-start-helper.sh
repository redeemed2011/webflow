#!/usr/bin/env bash

set -e
# set -x

# trap "exit" INT TERM ERR
# trap "kill 0" EXIT

echo "$(date) Generating ..."
go generate

# echo "$(date) Meta-linting ..."
# gometalinter ./cmd/${APP:?}/ ./internal/app/${APP:?}/ ./internal/pkg/**/ ./pkg/**/

echo "$(date) Testing ..."
go test -race -v .

echo
echo "$(date) Done!"
