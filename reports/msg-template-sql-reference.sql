-- src/data/msg_template.go 的 GORM / MySQL SQL 对照。
-- 2026-11-02 使用当前依赖的 GORM DryRun 核对 SQL 生成，未连接数据库。
-- 以下为独立语句示例，ID 7 是示例主键，不代表 INSERT 后实际生成的主键。
-- 实际代码使用 ? 和参数绑定，此处填入示例值方便阅读。

-- Find: db.Where("template_id = ?", templateID).First(&template)
SELECT *
FROM t_msg_template
WHERE template_id = 'tpl_demo'
ORDER BY id
LIMIT 1;

-- Create: db.Create(template)，假定 template.ID == 0。
-- id 为自增列；create_time / modify_time 在模型中只读，由 MySQL 维护。
INSERT INTO t_msg_template (
    template_id, rel_template_id, name, sign_name,
    source_id, channel, subject, content, status, ext
) VALUES (
    'tpl_demo', 'external_demo', '示例模板', NULL,
    'demo', 1, '测试标题', '测试内容', 1, 'demo-extra'
);

-- Save: Model + Where + Select + Updates。
-- Select 指定更新字段，包含零值；它在这里不是 SQL SELECT 查询。
UPDATE t_msg_template
SET rel_template_id = 'external_demo',
    name = '修改后的模板',
    sign_name = NULL,
    source_id = 'demo',
    channel = 1,
    subject = '',
    content = '测试内容',
    status = 1,
    ext = 'demo-extra'
WHERE id = 7 AND template_id = 'tpl_demo';

-- Save 的补充查询：仅 UPDATE 成功且 RowsAffected == 0 时执行。
-- count 为 0 表示已不存在；为 1 表示存在但可能未发生值变化。
SELECT COUNT(*)
FROM t_msg_template
WHERE id = 7 AND template_id = 'tpl_demo';

-- Delete: db.Where("template_id = ?", templateID).Delete(&MsgTemplate{})
-- 模型没有 gorm.DeletedAt，所以这里为物理删除。
DELETE FROM t_msg_template
WHERE template_id = 'tpl_demo';

-- .Error 和 .RowsAffected 读取执行结果，不额外生成 SQL。
-- TableName / gorm 字段标签用于映射，不会自行执行 CREATE TABLE。
-- 默认 GORM 写操作可带事务边界；以上仅列业务 SQL。
