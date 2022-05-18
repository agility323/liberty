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
GROUP="lserver-internal"
REPO="gate" # $CI_JOB_NAME
DOCKER_REPO="$GROUP/$REPO"
GITLAB_TAG=`date +"%Y%m%d_%H%M%S"`
IMAGE=${DOCKER_REGISTRY}/$DOCKER_REPO:$GITLAB_TAG

docker login -u "beihai" -p "123" $DOCKER_REGISTRY

echo "$ docker build -t $IMAGE"
docker build -t $REPO:$GITLAB_TAG -f ./Dockerfile -t $IMAGE .
echo "$ docker push $IMAGE"
docker push $IMAGE

docker logout $DOCKER_REGISTRY

