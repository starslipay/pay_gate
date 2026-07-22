#!/bin/bash
MODULE_NAME="pay_gate"
VERSION="v1.0.0"
IMAGE_NAME="${MODULE_NAME}:${VERSION}"

docker rm -f $MODULE_NAME
docker rmi -f $IMAGE_NAME
docker build -t $IMAGE_NAME .
docker run -d --name $MODULE_NAME --network dev_pay_net -p 30888:8888 $IMAGE_NAME
# docker run -d --name pay_gate --network dev_pay_net -p 30888:8080 pay_gate:v1.0.0
docker ps
docker logs $MODULE_NAME