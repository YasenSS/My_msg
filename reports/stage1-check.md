# 阶段 1 服务骨架验收

验收日期：2026-10-14（Asia/Shanghai）。

环境：`go version go1.26.4 darwin/arm64`；`go.mod` 要求 `go 1.26.4`。

## 构建

实际执行并成功退出：

```bash
GOCACHE=/private/tmp/my-msg-go-cache go build -o /private/tmp/my-msg-stage1 ./src
```

## 实测结果

使用临时 TOML 文件启动编译后的程序，通过本地 HTTP 请求检查状态码和 JSON 内容。每个端口检查后停止对应测试进程，再改写临时配置并启动下一次。项目中的 `config/config-test.toml` 保持 `8091`。

| 配置/请求 | 实际结果 |
| --- | --- |
| `port = 18091`，`GET /test` | HTTP 200，`{"test":"ok"}` |
| `port = 18091`，`GET /health` | HTTP 200，`{"status":"ok"}` |
| `port = 18092`，`GET /test` | HTTP 200，`{"test":"ok"}` |
| `port = 18092`，`GET /health` | HTTP 200，`{"status":"ok"}` |
| `port = 0` | 退出码 1，`初始化配置失败：端口必须在 1～65535 之间，当前值：0` |
| `port = 65536` | 退出码 1，`初始化配置失败：端口必须在 1～65535 之间，当前值：65536` |
| 配置文件不存在 | 退出码 1，`初始化配置失败：读取配置失败：open …: no such file or directory` |

## 手动复现

在项目根目录运行，先创建临时配置并启动服务：

```bash
printf '[common]\nport = 18091\n' > /private/tmp/my-msg-stage1-config.toml
/private/tmp/my-msg-stage1 --config /private/tmp/my-msg-stage1-config.toml
```

另一个终端执行：

```bash
curl -i http://127.0.0.1:18091/test
curl -i http://127.0.0.1:18091/health
```

用 Ctrl+C 停止服务，把临时配置的端口改为 `18092`，重启并请求对应新端口。然后依次验证错误配置，每次都检查退出码：

```bash
printf '[common]\nport = 0\n' > /private/tmp/my-msg-stage1-config.toml
/private/tmp/my-msg-stage1 --config /private/tmp/my-msg-stage1-config.toml
echo $?

printf '[common]\nport = 65536\n' > /private/tmp/my-msg-stage1-config.toml
/private/tmp/my-msg-stage1 --config /private/tmp/my-msg-stage1-config.toml
echo $?

/private/tmp/my-msg-stage1 --config /private/tmp/my-msg-stage1-missing/config.toml
echo $?
```

## 启动顺序与范围

当前顺序是参数解析 → 配置加载与校验 → Gin 初始化与路由注册 → HTTP 监听。暂无外部依赖，后续数据库初始化应放在配置校验之后、路由注册之前。

本次验证本地服务骨架；MySQL 和 Linux 部署尚未验证。
