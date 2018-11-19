#!/usr/bin/env bash

curl -o- https://raw.githubusercontent.com/creationix/nvm/v0.33.11/install.sh | bash
npm install -g nodemon


# Install the package manager.
go get -u github.com/golang/dep/cmd/dep

# Ensure all the packages are installed.
dep ensure


# Install the task runner.
# go get -u -v github.com/canthefason/go-watcher
go install github.com/canthefason/go-watcher/cmd/watcher
# go get -u -v github.com/go-task/task/cmd/task
# go get -u github.com/oxequa/realize


# Install the meta linter.
cd "${GOPATH}" && \
  curl -L https://git.io/vp6lP | sh
cd - || exit 1
