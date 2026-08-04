goctl api go -api pay_gate.api -dir .
goctl api swagger -api pay_gate.api -dir . -filename pay_gate
# 生成 Markdown 接口文档到根目录(--o 需为绝对路径, 用 $PSScriptRoot 取脚本所在目录)
goctl api doc --dir . --o $PSScriptRoot
