package model

import (
	"time"

	"gorm.io/gorm"
)

// Photo 图寻题目模型
type Photo struct {
	UserID      int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:投稿用户主键" json:"user_id"`
	ActivityID  int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:所属活动主键" json:"activity_id"`
	Title       string `gorm:"type:VARCHAR(128) NOT NULL;comment:图片标题" json:"title"`
	Description string `gorm:"type:TEXT;comment:图片描述" json:"description"`
	// 用 GCJ02 火星坐标系
	Latitude      float64 `gorm:"type:DECIMAL(10,7) NOT NULL;comment:图片纬度" json:"latitude"`
	Longitude     float64 `gorm:"type:DECIMAL(10,7) NOT NULL;comment:图片经度" json:"longitude"`
	CoordType     string  `gorm:"type:VARCHAR(16) DEFAULT 'gcj02' NOT NULL;comment:坐标系(wgs84/gcj02/bd09)" json:"coord_type"`
	Location      string  `gorm:"-:all"` // 忽略该字段的读写，只用于接收空间函数返回值
	ImageURL      string `gorm:"type:VARCHAR(512) NOT NULL;comment:原图URL" json:"image_url"`
	ThumbURL      string `gorm:"type:VARCHAR(512);comment:缩略图URL" json:"thumb_url"`
	ImageWidth    int    `gorm:"type:INT DEFAULT 0 NOT NULL;comment:原图宽度" json:"image_width"`
	ImageHeight   int    `gorm:"type:INT DEFAULT 0 NOT NULL;comment:原图高度" json:"image_height"`
	ThumbWidth    int    `gorm:"type:INT DEFAULT 0 NOT NULL;comment:缩略图宽度" json:"thumb_width"`
	ThumbHeight   int    `gorm:"type:INT DEFAULT 0 NOT NULL;comment:缩略图高度" json:"thumb_height"`
	Solved        bool    `gorm:"type:TINYINT(1) DEFAULT 0 NOT NULL;comment:是否已被破解" json:"solved"`
	SolvedCount   int     `gorm:"type:INT DEFAULT 0 NOT NULL;comment:破解成功次数" json:"solved_count"`
	AttemptsCount int     `gorm:"type:INT DEFAULT 0 NOT NULL;comment:答题次数" json:"attempts_count"`
	LikesCount    int     `gorm:"type:INT DEFAULT 0 NOT NULL;comment:点赞次数" json:"likes_count"`

	// 审核字段
	Status       string     `gorm:"type:VARCHAR(16) DEFAULT 'pending' NOT NULL;comment:审核状态(pending未审核/approved通过/rejected拒绝)" json:"status"`
	RejectReason string     `gorm:"type:VARCHAR(256);comment:拒绝原因" json:"reject_reason,omitempty"`
	ReviewedAt   *time.Time `gorm:"type:DATETIME(3);comment:审核时间" json:"reviewed_at,omitempty"`

	// 关联
	Activity Activity  `gorm:"foreignKey:ActivityID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Author   User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Comments []Comment `gorm:"foreignKey:PhotoID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Attempts []Attempt `gorm:"foreignKey:PhotoID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	BaseModel
}

// TableName 返回 Photo 对应的数据库表名
func (Photo) TableName() string {
	return "photo"
}

// BeforeCreate 创建前回调
func (info *Photo) BeforeCreate(_ *gorm.DB) error {
	return nil
}

// BeforeUpdate 更新前回调
func (info *Photo) BeforeUpdate(_ *gorm.DB) error {
	return nil
}

// AfterFind 查询后回调
func (info *Photo) AfterFind(_ *gorm.DB) error {
	return nil
}
