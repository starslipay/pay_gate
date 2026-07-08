# 定义镜像版本变量
$PAY_GATE_VERSION = "v1.0.4"

docker rmi -f pay_gate:$PAY_GATE_VERSION
docker build -t pay_gate:$PAY_GATE_VERSION .
docker save -o pay_gate.$PAY_GATE_VERSION.tar pay_gate:$PAY_GATE_VERSION

# 批量挂载
# multipass mount C:/study/starslipay worker1:/home/ubuntu/starslipay
# multipass mount C:/study/starslipay worker2:/home/ubuntu/starslipay
# multipass mount C:/study/starslipay worker3:/home/ubuntu/starslipay
# multipass mount C:/study/starslipay master1:/home/ubuntu/starslipay
# multipass mount C:/study/starslipay master2:/home/ubuntu/starslipay
# multipass mount C:/study/starslipay master3:/home/ubuntu/starslipay

# 所有需要导入镜像的虚拟机列表
$vmList = @("master1", "master2", "master3", "worker1", "worker2", "worker3")
# 虚拟机内镜像tar路径
$tarPath = "/home/ubuntu/starslipay/pay_gate/pay_gate.$PAY_GATE_VERSION.tar"

# 循环批量执行导入命令
foreach ($vm in $vmList) {
    Write-Host "==================== 正在向 $vm 导入镜像 ====================" -ForegroundColor Cyan
    multipass exec $vm -- sudo ctr -n k8s.io images import $tarPath
    # 判断单台执行结果
    if ($LASTEXITCODE -eq 0) {
        Write-Host "$vm 镜像导入成功`n" -ForegroundColor Green
    }
    else {
        Write-Host "$vm 镜像导入失败，请检查虚拟机或tar文件`n" -ForegroundColor Red
    }
}

Write-Host "全部节点导入任务执行完毕" -ForegroundColor Yellow
