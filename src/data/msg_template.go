package data

import (
	"time"

	"gorm.io/gorm"
)

type MsgTemplate struct {
	ID            int64 `gorm:"primaryKey;autoIncrement"`
	TemplateID    string
	RelTemplateID string  // 外部渠道的模板编号。
	Name          string  //模板名称
	SignName      *string // 短信签名，原项目将它传给短信供应商
	SourceID      string  // 业务来源编号
	Channel       *int
	Subject       string
	Content       string
	Status        *int
	Extra         *string   `gorm:"column:ext"` // Extra预留扩展字段；
	CreatedAt     time.Time `gorm:"column:create_time;->"`
	// column:create_time	将 CreatedAt 映射到数据库的 create_time 列
	// ->	只读：查询时可以填充字段，GORM 按模型新增、更新时不写入它
	UpdatedAt time.Time `gorm:"column:modify_time;->"`
}

//  *string、*int 是为了表达数据库的 NULL

func (MsgTemplate) TableName() string {
	return "t_msg_template"
}

var MsgTemplateNsp MsgTemplate

func (p *MsgTemplate) Find(db *gorm.DB, templateID string) (*MsgTemplate, error) {
	var template MsgTemplate
	err := db.Where("template_id = ?", templateID).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (p *MsgTemplate) Create(db *gorm.DB, template *MsgTemplate) error {
	if template == nil {
		return gorm.ErrInvalidData
	}
	return db.Create(template).Error
}
func (p *MsgTemplate) Save(db *gorm.DB, template *MsgTemplate) error {
	if template == nil || template.ID <= 0 || template.TemplateID == "" {
		return gorm.ErrInvalidData
	}
	res := db.Model(&MsgTemplate{}).
		Where("id = ? AND template_id = ?", template.ID, template.TemplateID).
		Select("RelTemplateID", "Name", "SignName",
			"SourceID", "Channel", "Subject", "Content",
			"Status", "Extra").
		Updates(template)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		return nil
	}

	var count int64
	err := db.Model(&MsgTemplate{}).
		Where("id = ? AND template_id = ?", template.ID, template.TemplateID).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (p *MsgTemplate) Delete(db *gorm.DB, templateID string) error {
	if templateID == "" {
		return gorm.ErrInvalidData
	}
	res := db.Where("template_id = ?", templateID).
		Delete(&MsgTemplate{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
