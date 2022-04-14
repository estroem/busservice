set -ex

rm -rf target
mkdir target

set +x

cp config.go target/config.go.backup

RABBITMQ_USERNAME=`kubectl get secret definition-default-user -o jsonpath='{.data.username}' | base64 --decode`
RABBITMQ_PASSWORD=`kubectl get secret definition-default-user -o jsonpath='{.data.password}' | base64 --decode`

sed -i '0,/%rabbitmq_username%/{s/%rabbitmq_username%/'$RABBITMQ_USERNAME'/}' config.go
sed -i '0,/%rabbitmq_password%/{s/%rabbitmq_password%/'$RABBITMQ_PASSWORD'/}' config.go

set -x

go build -o target/main main.go config.go

mv target/config.go.backup config.go

# image name
IMAGE=server-timing
VERSION=`cat VERSION`

docker build -t $IMAGE:$VERSION .
