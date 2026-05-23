package service

import (
	"io"
	"mime/multipart"
	"time"
	"tu-xun/common"
)

// ==================== 公共 ====================

// UserBrief 用户简要信息
type UserBrief struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ==================== Auth ====================

// Register:
// RegisterParams 注册参数
type RegisterParams struct {
	StudentID string `json:"student_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Password  string `json:"password" binding:"required,min=6,max=20"`
	Email     string `json:"email" binding:"required,email"`
}

// Login:
// LoginParams 登录参数
type LoginParams struct {
	StudentID string `json:"student_id" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

// ==================== Photo ====================

// Create:
// CreatePhotoParams 上传图片参数
type CreatePhotoParams struct {
	UserID         int64                 `form:"-"`
	Title          string                `form:"title" binding:"required"`
	Description    string                `form:"description"`
	LocationSecret string                `form:"location_secret" binding:"required"`
	ImageFile      *multipart.FileHeader `form:"image" binding:"required"`
}

// List:
// ListPhotoParams 图片列表查询参数
type ListPhotoParams struct {
	Page   int   `form:"page"`
	Limit  int   `form:"limit"`
	Solved *bool `form:"solved"`
}

// PhotoListItem 图片列表项
type PhotoListItem struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	ImageURL      string    `json:"image_url"`
	Author        UserBrief `json:"author"`
	Solved        bool      `json:"solved"`
	AttemptsCount int       `json:"attempts_count"`
	CreatedAt     time.Time `json:"created_at"`
}

// ListPhotosResponse 图片列表响应
type ListPhotosResponse struct {
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
	Items []PhotoListItem `json:"items"`
}

// GetByID:
// GetPhotoParams 获取图片详情参数
type GetPhotoParams struct {
	PhotoID       int64 `uri:"id" binding:"min=1"`
	CurrentUserID int64 `json:"-"`
}

// WinnerInfo 获奖者信息
type WinnerInfo struct {
	UserID    int64      `json:"user_id"`
	Name      string     `json:"name"`
	CreatedAt *time.Time `json:"created_at"`
}

// CurrentUserAttemptInfo 当前用户答题信息
type CurrentUserAttemptInfo struct {
	ID       int64  `json:"id"`
	Status   string `json:"status"`
	IsWinner bool   `json:"is_winner"`
}

// PhotoDetailResponse 图片详情响应
type PhotoDetailResponse struct {
	ID                 int64                   `json:"id"`
	Title              string                  `json:"title"`
	Description        string                  `json:"description"`
	ImageURL           string                  `json:"image_url"`
	Author             UserBrief               `json:"author"`
	Solved             bool                    `json:"solved"`
	AttemptsCount      int                     `json:"attempts_count"`
	CreatedAt          time.Time               `json:"created_at"`
	Winner             *WinnerInfo             `json:"winner,omitempty"`
	CurrentUserAttempt *CurrentUserAttemptInfo `json:"current_user_attempt,omitempty"`
}

// GetImageStream:
// ImageStream 封装图片流数据
type ImageStream struct {
	Reader      io.ReadCloser
	ContentType string
	Size        int64
	Filename    string
}

// ==================== Attempt ====================

// Create:
// CreateAttemptParams 提交答题参数
type CreateAttemptParams struct {
	PhotoID         int64                 `uri:"id" binding:"min=1"`
	UserID          int64                 `form:"-"`
	GuessedLocation string                `form:"guessed_location" binding:"required"`
	ImageFile       *multipart.FileHeader `form:"image" binding:"required"`
}

// SubmitAttemptResponse 提交答题响应
type SubmitAttemptResponse struct {
	AttemptID int64  `json:"attempt_id"`
	PhotoID   int64  `json:"photo_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// MyAttempts:
// MyAttemptsParams 获取我的答题记录参数
type MyAttemptsParams struct {
	PhotoID int64 `uri:"id" binding:"min=1"`
	UserID  int64 `form:"-"`
}

// AttemptItem 答题记录项
type AttemptItem struct {
	ID              int64      `json:"id"`
	ImageURL        string     `json:"image_url"`
	GuessedLocation string     `json:"guessed_location"`
	Status          string     `json:"status"`
	IsWinner        bool       `json:"is_winner"`
	ReviewedAt      *time.Time `json:"reviewed_at"`
}

// MyAttemptsResponse 我的答题记录响应
type MyAttemptsResponse struct {
	PhotoID    int64         `json:"photo_id"`
	Solved     bool          `json:"solved"`
	MyAttempts []AttemptItem `json:"my_attempts"`
}

// ==================== Admin ====================

// PendingPhotos:
// PendingPhotoForm 输入
type PendingPhotoParams struct {
	common.PagerForm
	status     string `form:"status" binding:"omitempty,oneof=pending approved rejected"`
	AdminLevel int    //审核员等级
}

// PendingPhotoForm 待审核图片项
type PendingPhotoForm struct {
	ID             int64  `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	LocationSecret string `json:"location_secret"`
	ThumbURL       string `json:"thumb_url"`
}

// PendingPhotosResponse 待审核图片列表响应
type PendingPhotosResponse struct {
	Total  int64              `json:"total"`
	Photos []PendingPhotoForm `json:"photos"`
}

// ReviewPhoto:
// ReviewPhotoParams 审核图片参数
type ReviewPhotoParams struct {
	PhotoID      int64  `uri:"id" binding:"min=1"`
	Action       string `json:"action" binding:"required"`
	RejectReason string `json:"reject_reason"`
	AdminLevel   int    //审核员等级
}

// ReviewPhotoResponse 审核图片响应
type ReviewPhotoResponse struct {
	ID      int64  `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// PendingAttempts:
// PendingAttemptForm 输入
type PendingAttemptParams struct {
	common.PagerForm
	status     string `form:"status" binding:"omitempty,oneof=pending approved rejected"`
	AdminLevel int    //审核员等级
}

// PendingAttemptForm 待审核答题项
type PendingAttemptForm struct {
	AttemptID       int64     `json:"attempt_id"`
	PhotoID         int64     `json:"photo_id"`
	PhotoTitle      string    `json:"photo_title"`
	ImageURL        string    `json:"image_url"`        //猜测照片
	ThumbURL        string    `json:"thumb_url"`        //原照片
	GuessedLocation string    `json:"guessed_location"` //猜测地址
	SubmittedAt     time.Time `json:"submitted_at"`
}

// PendingAttemptsResponse 待审核答题列表响应
type PendingAttemptsResponse struct {
	Total    int64                `json:"total"`
	Attempts []PendingAttemptForm `json:"items"`
}

// ReviewAttempt:
// ReviewAttemptParams 审核答题参数
type ReviewAttemptParams struct {
	AttemptID    int64  `uri:"id" binding:"min=1"`
	Action       string `json:"action" binding:"required"`
	RejectReason string `json:"reject_reason"`
}

// ReviewAttemptResponse 审核答题响应
type ReviewAttemptResponse struct {
	AttemptID   int64  `json:"attempt_id"`
	Status      string `json:"status"`
	IsWinner    bool   `json:"is_winner"`
	PhotoSolved bool   `json:"photo_solved"`
	Message     string `json:"message"`
}

// ClaimPrize:
// ClaimPrizeResponse 发放奖品响应
type ClaimPrizeResponse struct {
	PrizeID int64  `json:"prize_id"`
	Status  string `json:"status"`
}

// PendingComments:
// PendingCommentItem 待审核评论项
type PendingCommentItem struct {
	CommentID  int64     `json:"comment_id"`
	PhotoID    int64     `json:"photo_id"`
	PhotoTitle string    `json:"photo_title"`
	User       UserBrief `json:"user"`
	Comment    string    `json:"comment"`
	CreatedAt  time.Time `json:"created_at"`
}

// PendingCommentsResponse 待审核评论列表响应
type PendingCommentsResponse struct {
	Total int64                `json:"total"`
	Items []PendingCommentItem `json:"items"`
}

// ReviewComment:
// ReviewCommentParams 审核评论参数
type ReviewCommentParams struct {
	CommentID    int64  `uri:"id" binding:"min=1"`
	Action       string `json:"action" binding:"required"`
	RejectReason string `json:"reject_reason"`
}

// ReviewCommentResponse 审核评论响应
type ReviewCommentResponse struct {
	CommentID int64  `json:"comment_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// ==================== Prize ====================

// MyPrizes:
// PrizeItem 奖品项
type PrizeItem struct {
	ID         int64      `json:"id"`
	PhotoID    int64      `json:"photo_id"`
	PhotoTitle string     `json:"photo_title"`
	Status     string     `json:"status"`
	PrizeType  string     `json:"prize_type"`
	AwardedAt  *time.Time `json:"awarded_at"`
}

// MyPrizesResponse 我的奖品列表响应
type MyPrizesResponse struct {
	Prizes []PrizeItem `json:"prizes"`
}

// ==================== Comment ====================
// （暂无专用结构体，评论相关审核结构体见 Admin 分组）

// ==================== OSS ====================

// CreateBucket:
// CreateBucketRequest 创建 Bucket 请求
type CreateBucketRequest struct {
	BucketName string `json:"bucket_name" binding:"required"`
	Region     string `json:"region" binding:"required"`
}

// CreateBucketResponse 创建 Bucket 响应
type CreateBucketResponse struct {
	BucketName string `json:"bucket_name"`
	Location   string `json:"location"`
}
