# 阶段 2：模板 Find 方法验收

日期：2026-11-02（Asia/Shanghai）。证据来自用户提供的本机终端执行结果，并结合当前代码审阅；AI 未重复连接数据库执行。

## 执行命令

```bash
MY_MSG_TEST_CONFIG="$PWD/config/config-test.toml" \
go test -tags=integration ./src/data \
  -run '^TestMsgTemplateFind$' -v -count=1 -timeout=30s
```

## 结果

- `TestMsgTemplateFind/模板存在`：PASS，断言模板 ID、名称和 `ext` 字段映射。
- `TestMsgTemplateFind/模板不存在`：PASS，断言返回 `nil` 和 `gorm.ErrRecordNotFound`。
- 整体结果：`PASS`，`ok my-msg/src/data 0.622s`。
- 不存在模板时 GORM 打印 `record not found`，属于预期日志。
- 测试在事务中准备数据，通过 `t.Cleanup` 先回滚再关闭连接池，执行输出没有报告清理错误。

本次验收完成 Find 数据访问方法，不代表模板新增、修改、删除或 HTTP CRUD 已完成。

代码审阅发现测试文件尚未添加 `//go:build integration`；`-tags=integration` 本身不会限制无构建标签的文件，建议补齐标签，避免普通 `go test ./...` 也要求数据库配置。此问题不影响本次查询断言的通过结果。
