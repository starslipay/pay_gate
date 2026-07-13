#!/bin/bash

set -e

echo "=== 开始部署 etcd 集群 ==="

sudo kubectl apply -f namespace.yaml
sudo kubectl apply -f etcd-headless-svc.yaml
sudo kubectl apply -f etcd-statefulset.yaml
sudo kubectl apply -f etcd-configmap.yaml

echo -e "\n5. 查看 Pod 状态..."
sudo kubectl get pods -n pay-ns -o wide

echo -e "\n6. 查看 Service..."
sudo kubectl get svc -n pay-ns
