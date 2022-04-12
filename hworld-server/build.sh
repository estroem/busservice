set -ex

rm -rf target
mkdir target
go build -o target/main main.go

# image name
IMAGE=hworld-server
VERSION=`cat VERSION`

docker build -t $IMAGE:$VERSION .
