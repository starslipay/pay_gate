$MODULE_NAME = "user_mgr"
$VERSION = "v1.0.0"
$IMAGE_NAME = "${MODULE_NAME}:${VERSION}"

docker rm -f $MODULE_NAME
docker rmi -f $IMAGE_NAME
docker build -t $IMAGE_NAME .
docker run -d --name $MODULE_NAME --network local_deps_install_dev_net -p 30880:8080 $IMAGE_NAME
docker ps
docker logs $MODULE_NAME
