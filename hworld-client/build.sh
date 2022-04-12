set -ex

mkdir target
go build -o target/main main.go

# image name
IMAGE=hworld-client
VERSION=`cat VERSION`

docker build -t $IMAGE:$VERSION .
