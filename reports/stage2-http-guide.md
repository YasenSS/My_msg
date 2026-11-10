# 阶段 2 · 第 2 轮：模板 HTTP CRUD 教学

本文是供你亲自实现的成组示例，未写入当前项目业务源码。以下每个 Go 代码块都是标明路径文件的完整内容；新增文件时创建对应目录，已有文件按说明合并或替换。无需新增依赖，main.go 无需修改。

## 进度与本轮目标

2026-11-10 更新：Create / Save / Delete、TestMsgTemplateCRUD 与恢复后的独立 TestMsgTemplateFind 已在本地 MySQL 上实际运行通过，共用 newTestTx，见 reports/stage2-template-crud-check.md。下面 HTTP 代码仍是待用户实现的示例，数据层测试通过不代表 HTTP 接口验收通过。

本轮完整链路：路由 → 绑定请求 → 参数校验 → HandleInput → HandleProcess → GORM → MySQL → 统一 JSON 响应。

## 参考与修正

参考原项目（/Users/admin/Learning/msgcenter-main）：

- src/initialize/initialize.go：RegisterRouter，保留四个接口路径及方法。
- src/ctrl/ctrlmodel/model.go：保留 JSON 字段名、顶层 code/msg 与模板响应字段。
- src/ctrl/handler/hanlder.go：Run、HandlerIntf，继续采用输入校验与业务处理两步。
- src/ctrl/msg/*template.go：四个 Handler；纠正校验失败只设置 Code 却返回 nil、数据库错误被遗漏的问题。
- src/constant/error.go：保留本阶段使用的 0、8020、8021、8023，补齐原项目缺失的内部错误消息。
- src/data/constant.go：渠道 EMAIL=1 / SMS=2 / LARK=3；模板 PENDING=1 / NORMAL=2。

沿用原项目 HTTP 200 + 业务 code 的响应约定。0=成功；8020=参数无效（包括不存在的模板 ID）；8021=绑定失败（JSON 语法/类型不符等）；8023=内部错误。不新增未约定的业务码，数据库错误详情只写日志。不存在模板映射 8020 是本阶段明确的修正规则，原项目对此错误处理不完整。

本轮更正部分更新语义：请求缺字段就保留旧值；subject 显式传空字符串可清空；name/sourceID/content 不能清空；channel 只能取 1/2/3。更新请求用指针表达是否传入，JSON null 和未传一样按“不更新”处理。没有可更新字段直接拒绝。字段名及值类型保持原契约；长度和渠道合法性检查是本轮补充的校验。

## 文件职责与实现顺序

| 顺序 | 文件 | 作用 |
| --- | --- | --- |
| 1 | constant/error.go、data/constant.go | 错误码、渠道、状态 |
| 2 | ctrl/ctrlmodel/model.go | 请求绑定、响应字段 |
| 3 | ctrl/handler/handler.go | 校验失败立即终止处理 |
| 4 | ctrl/msg/template_common.go | 统一绑定、错误映射和响应 |
| 5 | ctrl/tools/template_id.go | 标准库生成 UUID v4 模板 ID |
| 6 | ctrl/msg 四个接口文件 | 新增、查询、修改、删除 |
| 7 | initialize/initialize.go | 注册原接口路径 |
| 8 | initialize/template_validation_test.go | 无数据库的 HTTP 参数校验测试 |

理解以下几个点后再写代码：

- ShouldBind 依据 HTTP 方法和 Content-Type 选择绑定器；GET 读取 query，POST JSON 请发送 application/json，也保留 form 标签。它还会执行 binding 标签校验。与自动写错误响应的 Bind 不同，ShouldBind 让我们统一返回业务格式。
- 绑定错误或 HandleInput 错误后不会进入 HandleProcess。每个请求使用自己独立的 Handler；任何校验失败都向上传递非 nil 错误。
- ctx 保存请求 context，WithContext 让数据库查询接收请求取消信号；这里只派生 GORM 会话，不新建连接池。连接池仍由 main 统一管理，Handler 不关闭它。本轮尚未加独立的 HTTP 查询时限。
- 更新中的“查询 → 合并字段 → 保存”包在 MySQL 事务中，SELECT FOR UPDATE 对存在的记录加锁，避免并发请求用旧模型覆盖别人的更新。事务函数返回 nil 时提交，返回错误时回滚；这只是模板更新需要的行锁，后面的消费者分布式锁仍属于阶段 6。
- valueOrZero 是泛型辅助函数：数据库里的 nil 签名/渠道输出为原响应契约中的空字符串/0，不改成 null。Req 中的指针是区分字段是否提供，和数据模型里用于表达 SQL NULL 的用途不同。
- Create 不接受用户指定状态，一律写入 PENDING。数据库操作成功后才向响应写 templateID。PENDING 表示待准备，NORMAL 表示可用于后续发送；当前没有审核接口，本阶段最后用 SQL 准备正常模板。
- tools.NewTemplateID 使用标准库生成 UUID v4，不引入原项目整套中间件工具包；版本位和变体位遵循 UUID v4 格式，随机源失败向上传递。

## 按文件实现

### `src/constant/error.go`

```go
package constant

import "errors"

const (
	SUCCESS           = 0
	ERR_INPUT_INVALID = 8020
	ERR_SHOULD_BIND   = 8021
	ERR_INTERNAL      = 8023
)

var ERR_HANDLE_INPUT = errors.New("handle input error")

func GetErrMsg(code int) string {
	switch code {
	case SUCCESS:
		return "ok"
	case ERR_INPUT_INVALID:
		return "input invalid"
	case ERR_SHOULD_BIND:
		return "should bind failed"
	default:
		return "internal error"
	}
}
```

### `src/data/constant.go`

```go
package data

type ChannelEnum int

const (
	Channel_EMAIL ChannelEnum = 1
	Channel_SMS   ChannelEnum = 2
	Channel_LARK  ChannelEnum = 3
)

type TemplateStatus int

const (
	TEMPLATE_STATUS_PENDING TemplateStatus = 1
	TEMPLATE_STATUS_NORMAL  TemplateStatus = 2
)
```

### `src/ctrl/ctrlmodel/model.go`

```go
package ctrlmodel

type RespComm struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type CreateTemplateReq struct {
	SourceID string `json:"sourceID" form:"sourceID" binding:"required,max=256"`
	Name     string `json:"name" form:"name" binding:"required,max=256"`
	Subject  string `json:"subject" form:"subject" binding:"required,max=256"`
	SignName string `json:"signName" form:"signName" binding:"max=256"`
	Channel  int    `json:"channel" form:"channel" binding:"oneof=1 2 3"`
	Content  string `json:"content" form:"content" binding:"required,max=4096"`
}

type CreateTemplateResp struct {
	RespComm
	TemplateID string `json:"templateID"`
}

type GetTemplateReq struct {
	TemplateID string `json:"templateID" form:"templateID" binding:"required,max=256"`
}

type GetTemplateResp struct {
	RespComm
	RelTemplateID string `json:"relTemplateID"`
	SourceID      string `json:"sourceID"`
	SignName      string `json:"signName"`
	Name          string `json:"name"`
	Subject       string `json:"subject"`
	Channel       int    `json:"channel"`
	Content       string `json:"content"`
}

type UpdateTemplateReq struct {
	TemplateID string  `json:"templateID" form:"templateID" binding:"required,max=256"`
	Name       *string `json:"name" form:"name" binding:"omitempty,min=1,max=256"`
	SourceID   *string `json:"sourceID" form:"sourceID" binding:"omitempty,min=1,max=256"`
	Subject    *string `json:"subject" form:"subject" binding:"omitempty,max=256"`
	Channel    *int    `json:"channel" form:"channel" binding:"omitempty,oneof=1 2 3"`
	Content    *string `json:"content" form:"content" binding:"omitempty,min=1,max=4096"`
}

type UpdateTemplateResp struct {
	RespComm
}

type DelTemplateReq struct {
	TemplateID string `json:"templateID" form:"templateID" binding:"required,max=256"`
}

type DelTemplateResp struct {
	RespComm
}
```

### `src/ctrl/handler/handler.go`

```go
package handler

type HandlerIntf interface {
	HandleInput() error
	HandleProcess() error
}

func Run(h HandlerIntf) error {
	if err := h.HandleInput(); err != nil {
		return err
	}
	return h.HandleProcess()
}
```

### `src/ctrl/msg/template_common.go`

```go
package msg

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"my-msg/src/constant"
	"my-msg/src/ctrl/ctrlmodel"
	"my-msg/src/ctrl/handler"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

func serveTemplate(
	c *gin.Context,
	req any,
	h handler.HandlerIntf,
	common *ctrlmodel.RespComm,
	resp any,
) {
	code := constant.SUCCESS
	if err := c.ShouldBind(req); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			code = constant.ERR_INPUT_INVALID
		} else {
			code = constant.ERR_SHOULD_BIND
		}
	} else if err := handler.Run(h); err != nil {
		switch {
		case errors.Is(err, constant.ERR_HANDLE_INPUT),
			errors.Is(err, gorm.ErrInvalidData),
			errors.Is(err, gorm.ErrRecordNotFound):
			code = constant.ERR_INPUT_INVALID
		default:
			code = constant.ERR_INTERNAL
			log.Printf("模板接口失败：path=%s err=%v", c.Request.URL.Path, err)
		}
	}
	common.Code = code
	common.Msg = constant.GetErrMsg(code)
	c.JSON(http.StatusOK, resp)
}

func blank(s string) bool {
	return strings.TrimSpace(s) == ""
}

func valueOrZero[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}
```

### `src/ctrl/tools/template_id.go`

```go
package tools

import (
	"crypto/rand"
	"fmt"
)

func NewTemplateID() (string, error) {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return "", fmt.Errorf("生成模板 ID 失败：%w", err)
	}
	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		id[0:4], id[4:6], id[6:8], id[8:10], id[10:16]), nil
}
```

### `src/ctrl/msg/create_template.go`

```go
package msg

import (
	"context"

	"my-msg/src/constant"
	"my-msg/src/ctrl/ctrlmodel"
	"my-msg/src/ctrl/tools"
	"my-msg/src/data"

	"github.com/gin-gonic/gin"
)

type CreateTemplateHandler struct {
	Req  ctrlmodel.CreateTemplateReq
	Resp ctrlmodel.CreateTemplateResp
	ctx  context.Context
}

func CreateTemplate(c *gin.Context) {
	h := CreateTemplateHandler{ctx: c.Request.Context()}
	serveTemplate(c, &h.Req, &h, &h.Resp.RespComm, &h.Resp)
}

func (h *CreateTemplateHandler) HandleInput() error {
	if blank(h.Req.Name) || blank(h.Req.SourceID) ||
		blank(h.Req.Content) || blank(h.Req.Subject) {
		return constant.ERR_HANDLE_INPUT
	}
	return nil
}

func (h *CreateTemplateHandler) HandleProcess() error {
	id, err := tools.NewTemplateID()
	if err != nil {
		return err
	}
	status := int(data.TEMPLATE_STATUS_PENDING)
	mt := &data.MsgTemplate{
		TemplateID: id,
		Name:       h.Req.Name,
		SourceID:   h.Req.SourceID,
		Subject:    h.Req.Subject,
		Content:    h.Req.Content,
		SignName:   &h.Req.SignName,
		Channel:    &h.Req.Channel,
		Status:     &status,
	}
	db := data.GetData().GetDB().WithContext(h.ctx)
	if err := data.MsgTemplateNsp.Create(db, mt); err != nil {
		return err
	}
	h.Resp.TemplateID = id
	return nil
}
```

### `src/ctrl/msg/get_template.go`

```go
package msg

import (
	"context"

	"my-msg/src/constant"
	"my-msg/src/ctrl/ctrlmodel"
	"my-msg/src/data"

	"github.com/gin-gonic/gin"
)

type GetTemplateHandler struct {
	Req  ctrlmodel.GetTemplateReq
	Resp ctrlmodel.GetTemplateResp
	ctx  context.Context
}

func GetTemplate(c *gin.Context) {
	h := GetTemplateHandler{ctx: c.Request.Context()}
	serveTemplate(c, &h.Req, &h, &h.Resp.RespComm, &h.Resp)
}

func (h *GetTemplateHandler) HandleInput() error {
	if blank(h.Req.TemplateID) {
		return constant.ERR_HANDLE_INPUT
	}
	return nil
}

func (h *GetTemplateHandler) HandleProcess() error {
	db := data.GetData().GetDB().WithContext(h.ctx)
	mt, err := data.MsgTemplateNsp.Find(db, h.Req.TemplateID)
	if err != nil {
		return err
	}
	h.Resp.RelTemplateID = mt.RelTemplateID
	h.Resp.SourceID = mt.SourceID
	h.Resp.SignName = valueOrZero(mt.SignName)
	h.Resp.Name = mt.Name
	h.Resp.Subject = mt.Subject
	h.Resp.Channel = valueOrZero(mt.Channel)
	h.Resp.Content = mt.Content
	return nil
}
```

### `src/ctrl/msg/update_template.go`

```go
package msg

import (
	"context"

	"my-msg/src/constant"
	"my-msg/src/ctrl/ctrlmodel"
	"my-msg/src/data"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UpdateTemplateHandler struct {
	Req  ctrlmodel.UpdateTemplateReq
	Resp ctrlmodel.UpdateTemplateResp
	ctx  context.Context
}

func UpdateTemplate(c *gin.Context) {
	h := UpdateTemplateHandler{ctx: c.Request.Context()}
	serveTemplate(c, &h.Req, &h, &h.Resp.RespComm, &h.Resp)
}

func (h *UpdateTemplateHandler) HandleInput() error {
	r := h.Req
	if blank(r.TemplateID) {
		return constant.ERR_HANDLE_INPUT
	}
	if r.Name == nil && r.SourceID == nil && r.Subject == nil &&
		r.Channel == nil && r.Content == nil {
		return constant.ERR_HANDLE_INPUT
	}
	for _, field := range []*string{r.Name, r.SourceID, r.Content} {
		if field != nil && blank(*field) {
			return constant.ERR_HANDLE_INPUT
		}
	}
	return nil
}

func (h *UpdateTemplateHandler) HandleProcess() error {
	db := data.GetData().GetDB().WithContext(h.ctx)
	return db.Transaction(func(tx *gorm.DB) error {
		// 锁住查到的记录，避免两个请求读旧数据后相互覆盖更新。
		mt, err := data.MsgTemplateNsp.Find(
			tx.Clauses(clause.Locking{Strength: "UPDATE"}), h.Req.TemplateID,
		)
		if err != nil {
			return err
		}
		if h.Req.Name != nil {
			mt.Name = *h.Req.Name
		}
		if h.Req.SourceID != nil {
			mt.SourceID = *h.Req.SourceID
		}
		if h.Req.Subject != nil {
			mt.Subject = *h.Req.Subject
		}
		if h.Req.Channel != nil {
			mt.Channel = h.Req.Channel
		}
		if h.Req.Content != nil {
			mt.Content = *h.Req.Content
		}
		return data.MsgTemplateNsp.Save(tx, mt)
	})
}
```

### `src/ctrl/msg/del_template.go`

```go
package msg

import (
	"context"

	"my-msg/src/constant"
	"my-msg/src/ctrl/ctrlmodel"
	"my-msg/src/data"

	"github.com/gin-gonic/gin"
)

type DelTemplateHandler struct {
	Req  ctrlmodel.DelTemplateReq
	Resp ctrlmodel.DelTemplateResp
	ctx  context.Context
}

func DelTemplate(c *gin.Context) {
	h := DelTemplateHandler{ctx: c.Request.Context()}
	serveTemplate(c, &h.Req, &h, &h.Resp.RespComm, &h.Resp)
}

func (h *DelTemplateHandler) HandleInput() error {
	if blank(h.Req.TemplateID) {
		return constant.ERR_HANDLE_INPUT
	}
	return nil
}

func (h *DelTemplateHandler) HandleProcess() error {
	db := data.GetData().GetDB().WithContext(h.ctx)
	return data.MsgTemplateNsp.Delete(db, h.Req.TemplateID)
}
```

### `src/initialize/initialize.go`

```go
package initialize

import (
	"net/http"

	"my-msg/src/ctrl/msg"

	"github.com/gin-gonic/gin"
)

func RegisterRouter(router *gin.Engine) {
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"test": "ok"})
	})
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.POST("/msg/create_template", msg.CreateTemplate)
	router.GET("/msg/get_template", msg.GetTemplate)
	router.POST("/msg/update_template", msg.UpdateTemplate)
	router.POST("/msg/del_template", msg.DelTemplate)
}
```

### `src/initialize/template_validation_test.go`

```go
package initialize

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"my-msg/src/constant"
	"my-msg/src/ctrl/ctrlmodel"

	"github.com/gin-gonic/gin"
)

func TestTemplateValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRouter(router)
	tests := []struct {
		name, method, path, body string
		code                     int
	}{
		{"创建缺参数", "POST", "/msg/create_template", `{}`, constant.ERR_INPUT_INVALID},
		{"JSON 格式错误", "POST", "/msg/create_template", `{`, constant.ERR_SHOULD_BIND},
		{"字段类型错误", "POST", "/msg/create_template", `{"channel":"email"}`, constant.ERR_SHOULD_BIND},
		{"非法渠道", "POST", "/msg/create_template", `{"name":"n","sourceID":"s","subject":"s","content":"c","channel":9}`, constant.ERR_INPUT_INVALID},
		{"空白名称", "POST", "/msg/create_template", `{"name":"  ","sourceID":"s","subject":"s","content":"c","channel":1}`, constant.ERR_INPUT_INVALID},
		{"查询缺ID", "GET", "/msg/get_template", "", constant.ERR_INPUT_INVALID},
		{"更新无字段", "POST", "/msg/update_template", `{"templateID":"demo"}`, constant.ERR_INPUT_INVALID},
		{"更新空名称", "POST", "/msg/update_template", `{"templateID":"demo","name":""}`, constant.ERR_INPUT_INVALID},
		{"更新非法渠道", "POST", "/msg/update_template", `{"templateID":"demo","channel":0}`, constant.ERR_INPUT_INVALID},
		{"删除缺ID", "POST", "/msg/del_template", `{}`, constant.ERR_INPUT_INVALID},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("HTTP=%d body=%s", w.Code, w.Body.String())
			}
			var got ctrlmodel.RespComm
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if got.Code != tc.code {
				t.Fatalf("预期 code=%d，实际：%+v", tc.code, got)
			}
		})
	}
}
```

## 验证命令与预期

### Find 独立测试参考

以下函数已恢复到 `src/data/msg_template_test.go`，无需再次添加。它覆盖 `ext` 映射和不存在的 ID，与 CRUD 生命周期测试互补。

```go
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
```

### 1. 原数据层测试

```bash
MY_MSG_TEST_CONFIG="$PWD/config/config-test.toml" \
go test -tags=integration ./src/data \
  -run '^TestMsgTemplate(Find|CRUD)$' -v -count=1 -timeout=30s
```

确认 Find 和 CRUD 测试都存在，再核对实际输出中的测试名称，不能仅看最后的 ok。预期全部 PASS。测试数据通过事务回滚，重复键和找不到模板的日志在对应失败场景下正常。

### 2. 不连接数据库的 HTTP 参数测试与构建

```bash
go test ./... -count=1 -timeout=30s
go build -o /private/tmp/my-msg-http ./src
```

普通测试会排除带 integration 标签的数据层测试。示例中的十个 HTTP 用例均在访问数据库前被拒绝，可检验绑定错误、参数错误和校验短路，不启动服务、不占端口。它们不能代替真实数据库下的正常请求联调。

### 3. 启动与完整 HTTP CRUD

```bash
/private/tmp/my-msg-http --config ./config/config-test.toml
```

另开终端创建模板：

```bash
curl -sS http://127.0.0.1:8091/msg/create_template \
  -H 'Content-Type: application/json' \
  -d '{"sourceID":"learning","name":"HTTP 测试模板","subject":"测试标题","signName":"","channel":1,"content":"你好，{{name}}"}'
```

预期响应形状（ID 为随机值）：

```json
{"code":0,"msg":"ok","templateID":"xxxxxxxx-xxxx-4xxx-xxxx-xxxxxxxxxxxx"}
```

把返回的 templateID 填入变量，然后依次执行以下命令，每一步检查响应中的 code：

```bash
template_id='这里替换为返回的模板ID'

curl -sS -G http://127.0.0.1:8091/msg/get_template \
  --data-urlencode "templateID=$template_id"

curl -sS http://127.0.0.1:8091/msg/update_template \
  -H 'Content-Type: application/json' \
  -d "{\"templateID\":\"$template_id\",\"name\":\"修改后的名称\",\"subject\":\"\"}"

curl -sS -G http://127.0.0.1:8091/msg/get_template \
  --data-urlencode "templateID=$template_id"

curl -sS http://127.0.0.1:8091/msg/del_template \
  -H 'Content-Type: application/json' \
  -d "{\"templateID\":\"$template_id\"}"

curl -sS -G http://127.0.0.1:8091/msg/get_template \
  --data-urlencode "templateID=$template_id"
```

预期依次为：查询成功 → 更新成功 → 查到新名称、空 subject、其他字段保留 → 删除成功 → 8020。HTTP 返回 200 本身不代表业务成功。这里的数据真正提交数据库，最后由删除接口清理；中途失败则保留模板 ID 方便排查和后续清理。

## 并发与当前边界

Save 内 UPDATE 后的 COUNT 是两条语句；本轮 HTTP 更新把它们放在同一事务内，并先锁定目标记录。删除仍使用唯一 template_id 条件进行物理删除。未来如果放开并发修改模板状态/渠道，需要讨论状态复核与发送中的模板一致性，不在本轮加入发送或缓存。

本轮沿用原项目当前开放的模板管理接口，没有增加鉴权；原项目读取但未实际使用的 UserId 不作为已具备鉴权的能力。签名、外部供应商模板 ID 的真实渠道规则留到对应阶段补充。

## 示例验证记录

2026-11-02：把本文 Go 代码块与当前 src/data、src/config、main.go 组合到 /private/tmp/my-msg-http-lesson/example 中，运行 gofmt、go test ./... -count=1 -timeout=30s、go build，以及数据集成测试的 -run '^$' 编译检查，均成功。十个 HTTP 参数用例实际运行通过。未连接数据库、未执行真实 HTTP CRUD，因此不能据此勾选本项目业务验收通过。

## 你这一轮要交付什么

按顺序实现尚未落地的 HTTP 文件，然后运行数据层集成测试、普通 HTTP 测试和编译，再跑一次 curl CRUD。反馈完整测试结果和接口响应；下一轮集中审阅、补异常联调及正常状态模板 SQL，完成阶段 2 验收。业务代码由你亲自实现，本文仅是教学示例。
