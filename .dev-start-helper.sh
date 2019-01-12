#!/usr/bin/env bash

set -e
# set -x

# trap "exit" INT TERM ERR
# trap "kill 0" EXIT

date
go generate

# date
# gometalinter ./cmd/${APP:?}/ ./internal/app/${APP:?}/ ./internal/pkg/**/ ./pkg/**/

date
go test -race -v .

echo
echo "Done testing!"
