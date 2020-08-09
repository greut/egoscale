#!/bin/sh

mkdir -p pkg/v2

for f in $(git ls-files internal/v2/*.go)
do
    git mv $f pkg/v2/
done

git ls-files *.go | \
    xargs sed -i 's|egoscale/internal|egoscale/pkg|g'
