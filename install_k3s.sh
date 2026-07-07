#!/bin/sh



# 1. 在 k3s 节点导入镜像
sudo ctr -n k8s.io images import /home/ubuntu/starslipay/pay_gate/pay_gate.tar

# 2. 触发 Deployment 更新（强制重启 Pod）
sudo kubectl rollout restart deployment pay-gate -n gateway-ns

# 3. 创建命名空间
sudo kubectl apply -f namespace.yaml

# 4. 创建 Deployment
sudo kubectl apply -f deployment.yaml

# 5.创建 Service
sudo kubectl apply -f service.yaml

# 6.创建 Ingress
sudo kubectl apply -f ingress.yaml

# 查看 Pod 状态
sudo kubectl get pods -n gateway-ns -o wide

# 查看 Service
sudo kubectl get svc -n gateway-ns

# 查看日志
sudo kubectl logs -n gateway-ns -l app=pay-gate -f