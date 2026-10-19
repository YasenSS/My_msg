# 阶段 2：MySQL 初始化验收

验收日期：2026-10-14（Asia/Shanghai）。

范围：配置校验、MySQL/GORM 连接池初始化、启动顺序和数据库初始化失败退出。模板表与 CRUD 尚未实现。

## 实际执行

```bash
GOCACHE=/private/tmp/my-msg-go-cache go build -o /private/tmp/my-msg-stage2-db ./src
```

构建成功，退出码 0。随后使用 Python 临时验证程序读取现有 TOML，创建权限为 0600 的临时配置，并运行以下形式的命令：

```bash
/private/tmp/my-msg-stage2-db --config <临时配置路径>
```

正确配置分别使用 HTTP 18091、18092；每次启动后请求 `/health` 和 `/test`，检查 HTTP 状态与 JSON，并检查日志中 MySQL 初始化成功先于路由注册。验证后只终止本次启动的测试进程。错误配置由编译后的程序实际执行，检查退出码为 1，且没有注册路由或报告 MySQL 初始化成功。

项目中的开发配置和密码未修改；没有写入业务数据。

## 实测结果

| 场景 | 结果 |
| --- | --- |
| HTTP 18091 | MySQL 初始化成功；/health 与 /test 均为 HTTP 200；数据库初始化先于路由注册 |
| HTTP 18092 | MySQL 初始化成功；/health 与 /test 均为 HTTP 200；数据库初始化先于路由注册 |
| HTTP port=0 | 退出码 1；2026/10/14 18:32:11 初始化配置失败：common.port 必须在 1 ~ 65535 之间 |
| HTTP port=65536 | 退出码 1；2026/10/14 18:32:11 初始化配置失败：common.port 必须在 1 ~ 65535 之间 |
| 错误 MySQL 密码 | 退出码 1；2026/10/14 18:32:11 初始化数据库失败：检查 MySQL 连接失败：Error 1045 (28000): Access denied for user 'my_msg_dev'@'localhost' (using password: YES) |
| 不存在的数据库 | 退出码 1；2026/10/14 18:32:11 初始化数据库失败：检查 MySQL 连接失败：Error 1044 (42000): Access denied for user 'my_msg_dev'@'localhost' to database 'missing_7cef58c7143f484a8d9c951a399e62ba' |
| 错误 MySQL 端口 | 退出码 1；2026/10/14 18:32:11 初始化数据库失败：检查 MySQL 连接失败：dial tcp 127.0.0.1:54416: connect: connection refused |
| max_open_conns=0 | 退出码 1；2026/10/14 18:32:11 初始化配置失败：MySQL 连接池参数不合法 |
| connect_timeout_seconds=0 | 退出码 1；2026/10/14 18:32:11 初始化配置失败：MySQL 连接超时必须大于 0,连接寿命不能为负数 |
| 配置文件不存在 | 退出码 1；明确报告读取配置失败；HTTP 未启动 |

## 手动复现

使用正确配置启动后，在另一终端请求两个测试接口：

```bash
/private/tmp/my-msg-stage2-db --config ./config/config-test.toml
curl -i http://127.0.0.1:8091/health
curl -i http://127.0.0.1:8091/test
```

停止测试服务，将开发配置复制到临时文件：

```bash
cp config/config-test.toml /private/tmp/my-msg-stage2-invalid.toml
chmod 600 /private/tmp/my-msg-stage2-invalid.toml
```

在副本中逐项修改密码、数据库名、MySQL 端口或校验参数，每次只修改一项；运行并检查退出码：

```bash
/private/tmp/my-msg-stage2-db --config /private/tmp/my-msg-stage2-invalid.toml
echo $?
```

预期错误配置退出码为 1，HTTP 服务没有启动。数据库名错误案例实际返回权限错误 1044；这证明错误数据库名阻止启动，不把它描述成实际返回了 Unknown database。

## 资源与范围

`NewData` 在初始化失败时关闭已创建的连接池；`run` 返回时通过 defer 关闭连接池。本次 HTTP 测试进程使用终止信号清理，不代表已实现或验证优雅退出；信号处理与在途任务等待留待阶段 6。

IPv4 地址拼接沿用当前的 `fmt.Sprintf` 实现。模板建表、模型与业务 CRUD 待完成，阶段 2 尚未完成。
