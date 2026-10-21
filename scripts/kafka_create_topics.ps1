# kafka_create_topics.ps1
# 在 Kafka 集群上创建项目所需的 topic (幂等: 已存在的 topic 自动跳过)
# 通过 docker exec 进入 kafka 容器, 调用 kafka-topics.sh
#
# 用法:
#   .\kafka_create_topics.ps1                                    # 创建默认的 3 个业务 topic
#   .\kafka_create_topics.ps1 -Topic my_topic                    # 创建单个自定义 topic
#   .\kafka_create_topics.ps1 -Topic t1,t2 -Partitions 6         # 批量创建, 指定分区数
#   .\kafka_create_topics.ps1 -ReplicationFactor 1               # 单节点环境(副本数=1)
#   .\kafka_create_topics.ps1 -ListOnly                          # 仅列出当前所有 topic
#
# 注意: 命令在 kafka 容器内部执行, 必须使用集群内部监听地址 kafka1:29092
#       (EXTERNAL 的 127.0.0.1:9092 只在宿主机可用, 容器内使用会拿到错误的 broker 元数据)

param(
    # 要创建的 topic 列表, 默认是 account_mgr 生产 / 两个 consumer 消费的 3 个业务 topic
    [string[]]$Topic = @(
        "topic_c2c_async_transfer",
        "topic_c2b_async_transfer",
        "t_update_balance"
    ),
    # 分区数(与 docker-compose 中 KAFKA_NUM_PARTITIONS 保持一致, 3 个消费者可并行消费)
    [int]$Partitions = 3,
    # 副本数(3 节点集群用 3, 单节点环境用 1)
    [int]$ReplicationFactor = 3,
    # 消息保留时长(毫秒), 默认 1 天
    [string]$RetentionMs = "86400000",
    # 单条消息最大字节数, 默认 10MB
    [string]$MaxMessageBytes = "10485760",
    # 执行命令的 kafka 容器
    [string]$Container = "kafka1",
    # 容器内部 bootstrap 地址
    [string]$Bootstrap = "kafka1:29092",
    # 仅列出 topic, 不做创建
    [switch]$ListOnly
)

$ErrorActionPreference = "Stop"

function Assert-KafkaContainer {
    $names = docker ps --filter "name=^/$Container$" --format "{{.Names}}" 2>$null
    if ($LASTEXITCODE -ne 0 -or $names -ne $Container) {
        Write-Host "错误: 容器 $Container 未运行, 请先启动 Kafka 集群" -ForegroundColor Red
        Write-Host "可用容器: docker ps" -ForegroundColor Yellow
        exit 1
    }
}

# 执行容器内的 kafka-topics.sh, 返回输出行数组
function Invoke-KafkaTopics {
    param([string]$Arguments)
    $out = docker exec $Container bash -c "/opt/kafka/bin/kafka-topics.sh $Arguments" 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "命令执行失败: kafka-topics.sh $Arguments" -ForegroundColor Red
        Write-Host ($out -join "`n") -ForegroundColor Red
        exit 1
    }
    return $out
}

Assert-KafkaContainer

# ---------- 仅列出 ----------
if ($ListOnly) {
    Write-Host "`n--- 当前所有 topic ---" -ForegroundColor Cyan
    Invoke-KafkaTopics "--bootstrap-server $Bootstrap --list" | ForEach-Object { Write-Host "  $_" }
    exit 0
}

# ---------- 读取已存在的 topic ----------
$existing = @(Invoke-KafkaTopics "--bootstrap-server $Bootstrap --list" | ForEach-Object { $_.Trim() } | Where-Object { $_ -ne "" })

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "Bootstrap:         $Bootstrap"
Write-Host "Partitions:        $Partitions"
Write-Host "ReplicationFactor: $ReplicationFactor"
Write-Host "RetentionMs:       $RetentionMs"
Write-Host "MaxMessageBytes:   $MaxMessageBytes"
Write-Host "Topic 数:          $($Topic.Count)"
Write-Host "========================================" -ForegroundColor Cyan

$created = @()
$skipped = @()

foreach ($name in $Topic) {
    if ($existing -contains $name) {
        Write-Host "  [跳过] $name 已存在" -ForegroundColor Yellow
        $skipped += $name
        continue
    }

    $cmdArgs = "--bootstrap-server $Bootstrap --create --if-not-exists " +
               "--topic $name " +
               "--partitions $Partitions " +
               "--replication-factor $ReplicationFactor " +
               "--config retention.ms=$RetentionMs " +
               "--config max.message.bytes=$MaxMessageBytes"
    Write-Host "  [创建] $name ..." -ForegroundColor Green
    Invoke-KafkaTopics $cmdArgs | Out-Null
    $created += $name
}

# ---------- 输出创建结果详情 ----------
if ($created.Count -gt 0) {
    Write-Host "`n--- 已创建 topic 详情 ---" -ForegroundColor Cyan
    foreach ($name in $created) {
        Invoke-KafkaTopics "--bootstrap-server $Bootstrap --describe --topic $name" |
            ForEach-Object { Write-Host "  $_" }
    }
}

# ---------- 汇总 ----------
Write-Host "`n--- 汇总 ---" -ForegroundColor Cyan
Write-Host "  新建: $($created.Count) 个 $(if ($created.Count -gt 0) { '(' + ($created -join ', ') + ')' })" -ForegroundColor Green
Write-Host "  跳过: $($skipped.Count) 个 $(if ($skipped.Count -gt 0) { '(' + ($skipped -join ', ') + ')' })" -ForegroundColor Yellow
