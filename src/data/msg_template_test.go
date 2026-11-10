//go:build integration

package data

import (
	"crypto/rand"
	"errors"
	"my-msg/src/config"
	"os"
	"testing"

	mysqldriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

func newTestTx(t *testing.T) *gorm.DB {
	t.Helper()

	configPath := os.Getenv("MY_MSG_TEST_CONFIG")
	if configPath == "" {
		t.Fatal("请设置 MY_MSG_TEST_CONFIG")
	}
	if err := config.Init(configPath); err != nil {
		t.Fatal(err)
	}

	d, err := NewData(config.Conf)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := d.Close(); err != nil {
			t.Errorf("关闭连接池失败：%v", err)
		}
	})

	tx := d.GetDB().Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	t.Cleanup(func() {
		if err := tx.Rollback().Error; err != nil {
			t.Errorf("回滚失败：%v", err)
		}
	})
	return tx
}
func TestMsgTemplateCRUD(t *testing.T) {
	tx := newTestTx(t)

	dt := &MsgTemplate{
		TemplateID: "test_crud_" + rand.Text(),
		Name:       "原始名称",
		SourceID:   "test",
		Subject:    "原始标题",
		Content:    "测试内容",
	}

	// 新增成功，并回填主键。
	if err := MsgTemplateNsp.Create(tx, dt); err != nil {
		t.Fatalf("新增失败：%v", err)
	}
	if dt.ID <= 0 {
		t.Fatal("新增后没有回填主键")
	}

	got, err := MsgTemplateNsp.Find(tx, dt.TemplateID)
	if err != nil {
		t.Fatalf("新增后查询失败：%v", err)
	}
	if got.ID != dt.ID || got.Name != dt.Name {
		t.Fatalf("新增后的数据不符合预期：%+v", got)
	}

	// 使用新主键、相同模板 ID，验证模板唯一约束。
	duplicate := *dt
	duplicate.ID = 0
	err = MsgTemplateNsp.Create(tx, &duplicate)

	var mysqlErr *mysqldriver.MySQLError
	if !errors.As(err, &mysqlErr) || mysqlErr.Number != 1062 {
		t.Fatalf("预期重复键错误 1062，实际：%v", err)
	}

	// 修改普通字段，并验证空字符串能够保存。
	got.Name = "修改后的名称"
	got.Subject = ""
	if err := MsgTemplateNsp.Save(tx, got); err != nil {
		t.Fatalf("更新失败：%v", err)
	}

	got, err = MsgTemplateNsp.Find(tx, dt.TemplateID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "修改后的名称" || got.Subject != "" {
		t.Fatalf("更新后的数据不符合预期：%+v", got)
	}

	// 内容不变的更新也应该成功。
	if err := MsgTemplateNsp.Save(tx, got); err != nil {
		t.Fatalf("无变化更新失败：%v", err)
	}

	if err := MsgTemplateNsp.Delete(tx, dt.TemplateID); err != nil {
		t.Fatalf("删除失败：%v", err)
	}
	if _, err := MsgTemplateNsp.Find(tx, dt.TemplateID); !errors.Is(
		err, gorm.ErrRecordNotFound,
	) {
		t.Fatalf("删除后应查不到模板，实际：%v", err)
	}

	// 重复删除应明确返回不存在。
	if err := MsgTemplateNsp.Delete(tx, dt.TemplateID); !errors.Is(
		err, gorm.ErrRecordNotFound,
	) {
		t.Fatalf("重复删除应返回不存在，实际：%v", err)
	}

	// 保存已删除的模板，应失败，不能重新插入。
	if err := MsgTemplateNsp.Save(tx, got); !errors.Is(
		err, gorm.ErrRecordNotFound,
	) {
		t.Fatalf("保存已删除模板应返回不存在，实际：%v", err)
	}
	if _, err := MsgTemplateNsp.Find(tx, dt.TemplateID); !errors.Is(
		err, gorm.ErrRecordNotFound,
	) {
		t.Fatalf("Save 不应重新插入已删除模板，实际：%v", err)
	}
}

func TestMsgTemplateFind(t *testing.T) {
	tx := newTestTx(t)
	id := "test_find_" + rand.Text()
	err := tx.Exec(`
		INSERT INTO t_msg_template
		(template_id, rel_template_id, name, source_id, subject, content, ext)
		VALUES (?, '', '测试模板', 'test', '测试标题', '测试内容', ?)
	`, id, "test-extra").Error
	if err != nil {
		t.Fatal(err)
	}
	t.Run("模板存在", func(t *testing.T) {
		got, err := MsgTemplateNsp.Find(tx, id)
		if err != nil || got == nil {
			t.Fatalf("查询失败：got=%+v err=%v", got, err)
		}
		if got.TemplateID != id || got.Name != "测试模板" ||
			got.Extra == nil || *got.Extra != "test-extra" {
			t.Fatalf("模板或 ext 映射错误：%+v", got)
		}
	})
	t.Run("模板不存在", func(t *testing.T) {
		got, err := MsgTemplateNsp.Find(tx, id+"_missing")
		if got != nil || !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("预期 nil 和 ErrRecordNotFound，实际：got=%+v err=%v", got, err)
		}
	})
}
