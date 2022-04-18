set -ex

mkdir target

set +x

cp config.go target/config.go.backup

trap "mv target/config.go.backup config.go; rm -rf target" EXIT

RABBITMQ_USERNAME=`kubectl get secret definition-default-user -o jsonpath='{.data.username}' | base64 --decode`
RABBITMQ_PASSWORD=`kubectl get secret definition-default-user -o jsonpath='{.data.password}' | base64 --decode`

sed -i '0,/%rabbitmq_username%/{s/%rabbitmq_username%/'$RABBITMQ_USERNAME'/}' config.go
sed -i '0,/%rabbitmq_password%/{s/%rabbitmq_password%/'$RABBITMQ_PASSWORD'/}' config.go

set -x

go build -o target ./...

IMAGE=`grep module go.mod | head -n 1 | sed 's/module //g'`
VERSION=`cat VERSION`

docker build --build-arg module_name=$IMAGE -t $IMAGE:$VERSION .
