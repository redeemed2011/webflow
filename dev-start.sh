#!/usr/bin/env bash

set -e
set -x

# Ensure all the packages are installed and up to date.
dep ensure -update
# go mod download
# go mod tidy
# go mod vendor

# Run the project every time a file changes.
# watcher -depth 6 ./.dev-start-helper.sh
nodemon \
  --watch . \
  --ext '.go' \
  --ignore 'mock' \
  --verbose \
  --exec "${SHELL}" ./.dev-start-helper.sh
