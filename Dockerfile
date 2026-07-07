# 先删除旧镜像，不存在也不会报错 docker rmi -f pay_gate
# 构建镜像 docker build -t pay_gate .
# 导出镜像 docker save -o pay_gate.tar pay_gate:latest
# 多阶段构建减小镜像体积
FROM golang:1.25-alpine AS builder
WORKDIR /app
# 新增国内GOPROXY配置
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download
COPY . .
# 编译网关
RUN go build -o pay_gate .

# 运行镜像
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/pay_gate .
COPY --from=builder /app/etc ./etc
# 时区
RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai
EXPOSE 8888
CMD ["./pay_gate", "-f", "./etc/paygate.yaml"]
