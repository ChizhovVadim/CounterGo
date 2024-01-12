#!/bin/bash

versionName="5.5"
gitRevision=$(git rev-list -1 HEAD)
buildDate=$(date +"%Y-%m-%d")

# make goos goarch extension
function make {
        GOOS=$1 GOARCH=$2 go build \
                -ldflags "-X 'main.gitRevision=$gitRevision' -X 'main.buildDate=$buildDate' -X 'main.versionName=$versionName'" \
                -o counter-$versionName-$1-$2$3 \
                -tags=avx,embed \
                ./cmd/counter
}

#make  darwin amd64 ""
#make  darwin arm64 ""
make   linux amd64 ""
make windows amd64 ".exe"
