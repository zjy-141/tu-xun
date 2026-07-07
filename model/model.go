package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BaseModel struct {
	ID        int64          `gorm:"primaryKey;UNSIGNED;NOT NULL;comment:主键" json:"id"`
	CreatedAt time.Time      `gorm:"type:DATETIME(3);NOT NULL;comment:创建时间" json:"-"`
	UpdatedAt time.Time      `gorm:"type:DATETIME(3);NOT NULL;comment:更新时间" json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"type:DATETIME(3);NULL;index;comment:删除时间" json:"-"`
}

type Fields json.RawMessage

// GormDataType 返回 Fields 在数据库中的存储类型
func (n Fields) GormDataType() string {
	return "TEXT"
}

// GormValue 将 Fields 序列化为数据库可存储的值
func (n Fields) GormValue(_ context.Context, _ *gorm.DB) clause.Expr {
	if len(n) == 0 {
		return clause.Expr{SQL: "?", Vars: []any{"null"}}
	}
	return clause.Expr{SQL: "?", Vars: []any{string(n)}}
}

// Scan 从数据库读取值到 Fields
func (n *Fields) Scan(value any) error {
	*n = []byte(fmt.Sprintf("%s", value))
	return nil
}

// MarshalJSON 将 Fields 序列化为 JSON（空时返回 null）
func (n Fields) MarshalJSON() ([]byte, error) {
	if len(n) == 0 {
		return []byte("null"), nil
	}
	return n, nil
}

// UnmarshalJSON 从 JSON 反序列化到 Fields
func (n *Fields) UnmarshalJSON(resp []byte) error {
	if n == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	*n = append((*n)[0:0], resp...)
	return nil
}
