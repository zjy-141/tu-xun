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
	// 审核字段
	Status       string     `gorm:"type:VARCHAR(16) DEFAULT 'pending' NOT NULL;comment:审核状态(pending未审核/approved通过/rejected拒绝)" json:"status"`
	RejectReason string     `gorm:"type:VARCHAR(256);comment:拒绝原因" json:"reject_reason,omitempty"`
	ReviewedAt   *time.Time `gorm:"type:DATETIME(3);comment:审核时间" json:"reviewed_at,omitempty"`
}

type Fields json.RawMessage

func (n Fields) GormDataType() string {
	return "TEXT"
}

func (n Fields) GormValue(_ context.Context, _ *gorm.DB) clause.Expr {
	if len(n) == 0 {
		return clause.Expr{SQL: "?", Vars: []any{"null"}}
	}
	return clause.Expr{SQL: "?", Vars: []any{string(n)}}
}

func (n *Fields) Scan(value any) error {
	*n = []byte(fmt.Sprintf("%s", value))
	return nil
}

func (n Fields) MarshalJSON() ([]byte, error) {
	if len(n) == 0 {
		return []byte("null"), nil
	}
	return n, nil
}

func (n *Fields) UnmarshalJSON(resp []byte) error {
	if n == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	*n = append((*n)[0:0], resp...)
	return nil
}
