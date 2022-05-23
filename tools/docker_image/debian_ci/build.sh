#!/bin/bash

DOCKER_REGISTRY="10.1.71.45:5000"
NAMESPACE="base"
APP="debian_ci"
TAG="latest"
IMAGE=${DOCKER_REGISTRY}/$NAMESPACE/$APP:$TAG

docker rmi $IMAGE
docker build -t $APP:$TAG -f ./Dockerfile -t $IMAGE .
docker login -u "lserver" -p "lserver" $DOCKER_REGISTRY
docker push $IMAGE
docker logout $DOCKER_REGISTRY
