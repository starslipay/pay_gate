// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package main

import (
	"flag"
	"fmt"

	"github.com/starslipay/pay_gate/internal/config"
	"github.com/starslipay/pay_gate/internal/handler"
	"github.com/starslipay/pay_gate/internal/middleware"
	_ "github.com/starslipay/pay_gate/internal/response"
	"github.com/starslipay/pay_gate/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/paygate.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)

	// 全局中间件: 把请求路径注入 ctx, 供 metrics 打点使用
	server.Use(middleware.MetricMethodMiddleware)
	// 全局中间件: 接口维度令牌桶限流(基于 Redis, 多网关实例全局共享配额)
	server.Use(ctx.RateLimiter)

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

/*
echo "=== 节点信息 ==="
sudo kubectl get nodes -o wide

echo -e "\n=== 服务信息 ==="
sudo kubectl get svc -n pay-ns

echo -e "\n=== Pod 状态 ==="
sudo kubectl get pods -n pay-ns -o wide
*/
