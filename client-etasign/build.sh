set -ex

mkdir target

trap "rm -rf target" EXIT

go build -o target ./...

IMAGE=`grep module go.mod | head -n 1 | sed 's/module //g'`
VERSION=`cat VERSION`

docker build -t $IMAGE:$VERSION .