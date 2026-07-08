#!/bin/sh

PAY_GATE_VERSION=v1.0.2

echo "=== 开始部署 pay_gate ${PAY_GATE_VERSION} ==="

echo -e "\n1. 导入镜像..."
sudo ctr -n k8s.io images import /home/ubuntu/starslipay/pay_gate/pay_gate.${PAY_GATE_VERSION}.tar

echo -e "\n2. 应用配置文件..."
sudo kubectl apply -f namespace.yaml
sudo kubectl apply -f deployment.yaml
sudo kubectl apply -f service.yaml
sudo kubectl apply -f ingress.yaml

echo -e "\n3. 等待部署完成..."
sudo kubectl rollout status deployment pay-gate -n gateway-ns --timeout=120s

echo -e "\n=== 部署完成 ==="
echo -e "\nPod 状态："
sudo kubectl get pods -n gateway-ns -o wide

echo -e "\nService 状态："
sudo kubectl get svc -n gateway-ns

echo -e "\n最近日志："
sudo kubectl logs -n gateway-ns -l app=pay-gate --tail=20