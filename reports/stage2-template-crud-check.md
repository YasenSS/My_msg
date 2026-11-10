# 阶段 2：模板数据层复验

日期：2026-11-10（Asia/Shanghai）。环境：Go 1.26.4，macOS arm64，本地 MySQL。

## 执行与结果

在本次 PR worktree 内实际执行：

```sh
go test ./... -count=1 -timeout=30s
go vet ./...
go build -o /private/tmp/my-msg-pr-build ./src
MY_MSG_TEST_CONFIG=/Users/admin/Learning/my-msg/config/config-test.toml \
  go test -tags=integration ./src/data -v -count=1 -timeout=30s
```

- 普通测试命令成功，各包当前没有不带 integration 标签的测试。
- go vet 与构建成功。
- 使用临时本地配置启动构建产物，/health 与 /test 返回预期 HTTP 200 JSON；HTTP 端口为 0 时退出码为 1。只终止本次创建的测试进程，未验证优雅退出。
- TestMsgTemplateCRUD：PASS。验证新增并回填主键、读取、唯一键冲突、普通字段与空字符串更新、无变化更新、删除、重复删除及 Save 不会重新插入已删除模板。
- TestMsgTemplateFind：PASS，模板存在和模板不存在两个子测试均通过，覆盖 ext 字段映射。
- 集成测试整体结果：PASS，ok my-msg/src/data 0.601s。
- 重复键 1062 和 record not found 为预期场景；事务回滚和连接池关闭未报告错误。

本次 PR 从 main 建立独立分支，迁入当前重写项目快照，不改写原 master 历史。
本机开发配置留在原工作区，PR 仅提供不含密码的 config/config.example.toml，并忽略 config/config-test.toml。

模板 HTTP CRUD 仍仅是教学示例，真实 HTTP CRUD、Linux 运行与部署尚未验证。阶段 2 整体验收未完成。
