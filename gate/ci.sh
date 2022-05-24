#!/bin/bash

# build exe
go version
go build -o gate
if [ $? -ne 0 ]; then
	echo  "Error: go build fail"
	exit 1
fi

# build image
DOCKER_REGISTRY="10.1.71.45:5000"
NAMESPACE="lserver-internal"
APP="gate"
TAG="latest" # `date +"%Y%m%d_%H%M%S"`
DOCKER_REPO="$NAMESPACE/$APP"
IMAGE=${DOCKER_REGISTRY}/$DOCKER_REPO:$TAG

docker login -u "xxx" -p "123" $DOCKER_REGISTRY

echo "$ docker rmi $IMAGE"
docker rmi $IMAGE
echo "$ docker build -t $APP:$TAG -f ./Dockerfile -t $IMAGE ."
docker build -t $APP:$TAG -f ./Dockerfile -t $IMAGE .
echo "$ docker push $IMAGE"
docker push $IMAGE

docker logout $DOCKER_REGISTRY

