$VERSION = "v1.0.0"

pushd ..

# 先删容器，避免名称冲突
docker rm -f pay_gate
docker rmi -f pay_gate:$VERSION
docker build -t pay_gate:$VERSION .
docker run -d --name pay_gate --network local_deps_install_dev_net -p 30888:8888 pay_gate:$VERSION
docker ps
docker logs pay_gate -f

popd