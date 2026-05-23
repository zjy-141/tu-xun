package service

import (
	"io"
	"mime/multipart"
)

// ==================== aliyunOSS 参数 ====================
type CreateBucketRequest struct {
	BucketName string `json:"bucket_name" binding:"required"`
	Region     string `json:"region" binding:"required"`
	// ... 其他可选参数，如存储类型、ACL等
}

type CreateBucketResponse struct {
	BucketName string `json:"bucket_name"`
	Location   string `json:"location"`
	// ... 其他返回信息
}

// ==================== Story 参数 ====================

// CreateStoryParams 发布故事参数
type CreateStoryParams struct {
	PhotoID  int64  `uri:"id" binding:"min=1"`
	UserID   int64  `json:"-"`
	Content  string `json:"content" binding:"required"`
	MediaURL string `json:"media_url"`
}

// ListStoryByPhotoParams 获取图片下故事列表参数
type ListStoryByPhotoParams struct {
	PhotoID int64 `uri:"id" binding:"min=1"`
}

// ==================== Attempt 参数 ====================

// CreateAttemptParams 提交答题参数
type CreateAttemptParams struct {
	PhotoID         int64                 `uri:"id" binding:"min=1"`
	UserID          int64                 `form:"-"`
	GuessedLocation string                `form:"guessed_location" binding:"required"`
	ImageFile       *multipart.FileHeader `form:"image" binding:"required"`
}

// MyAttemptsParams 获取我的答题记录参数
type MyAttemptsParams struct {
	PhotoID int64 `uri:"id" binding:"min=1"`
	UserID  int64 `form:"-"`
}

// ==================== Photo 参数 ====================

// ImageStream 封装图片流数据
type ImageStream struct {
	Reader      io.ReadCloser
	ContentType string
	Size        int64
	Filename    string
}

// CreatePhotoParams 上传图片参数
type CreatePhotoParams struct {
	UserID         int64                 `form:"-"`
	Title          string                `form:"title" binding:"required"`
	Description    string                `form:"description"`
	LocationSecret string                `form:"location_secret" binding:"required"`
	ImageFile      *multipart.FileHeader `form:"image" binding:"required"`
}

// ListPhotoParams 图片列表查询参数
type ListPhotoParams struct {
	Page   int   `form:"page"`
	Limit  int   `form:"limit"`
	Solved *bool `form:"solved"`
}

// GetPhotoParams 获取图片详情参数
type GetPhotoParams struct {
	PhotoID       int64 `uri:"id" binding:"min=1"`
	CurrentUserID int64 `json:"-"`
}

// ==================== Auth 参数 ====================

// RegisterParams 注册参数
type RegisterParams struct {
	StudentID string `json:"student_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Password  string `json:"password" binding:"required,min=6,max=20"`
	Email     string `json:"email" binding:"required,email"`
}

// LoginParams 登录参数
type LoginParams struct {
	StudentID string `json:"student_id" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

// ==================== Admin 参数 ====================

// PendingPhotosParams 待审核图片列表参数
type PendingPhotosParams struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

// ReviewPhotoParams 审核图片参数
type ReviewPhotoParams struct {
	PhotoID      int64  `uri:"id" binding:"min=1"`
	Action       string `json:"action" binding:"required"`
	RejectReason string `json:"reject_reason"`
}

// PendingAttemptsParams 待审核答题列表参数
type PendingAttemptsParams struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

// ReviewAttemptParams 审核答题参数
type ReviewAttemptParams struct {
	AttemptID    int64  `uri:"id" binding:"min=1"`
	Action       string `json:"action" binding:"required"`
	RejectReason string `json:"reject_reason"`
}
