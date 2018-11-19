#!/usr/bin/env sh

set -e
set -x


# Run the project every time a file changes.
watcher -depth 6 ./dev-start-helper.sh
