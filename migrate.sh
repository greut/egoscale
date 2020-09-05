#!/bin/sh

mkdir -p pkg/v2

for f in $(git ls-files internal/v2/*.go)
do
    git mv $f pkg/v2/
done

git ls-files *.go | \
    xargs sed -i 's|egoscale/internal|egoscale/pkg|g'

sed -i "s|v2\\( *\\*v2.ClientWithResponses\\)|V2\\1|g" client.go
git ls-files *.go | \
    xargs sed -i "s|\\.v2|.V2|g"
