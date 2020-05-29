#!/bin/bash

versionName="3.6dev"
gitRevision=$(git rev-list -1 HEAD)
buildDate=$(date +"%Y-%m-%d")

# make goos goarch extension
function make {
        GOOS=$1 GOARCH=$2 go build \
                -ldflags "-X 'main.gitRevision=$gitRevision' -X 'main.buildDate=$buildDate' -X 'main.versionName=$versionName'" \
                -o counter-$versionName-$1-$2$3 \
                github.com/ChizhovVadim/CounterGo/counter
}

make  darwin amd64 ""
make   linux amd64 ""
make windows amd64 ".exe"
