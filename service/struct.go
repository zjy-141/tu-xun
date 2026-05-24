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
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}
type UserForm struct {
	ID        int64  `json:"-"`
	StudentID string `json:"student_id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Level     int    `json:"level"`
	QQ        string `json:"qq"`
	WeiXin    string `json:"weixin"`
}

// PhotoListItem 图片列表项
type PhotoForm struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	ThumbURL      string    `json:"thumb_url"`
	Author        UserBrief `json:"author"`
	Solved        bool      `json:"solved"`
	CreatedAt     time.Time `json:"created_at"`
	AttemptsCount int       `json:"attempts_count"`
	LikesCount    int       `json:"likes_count"`
}
type AttemptForm struct {
	ID              int64     `json:"id"`
	ImageURL        string    `json:"image_url"`
	CommentText     string    `json:"comment,omitempty"`
	GuessedLocation string    `json:"guessed_location"`
	CreatedAt       time.Time `json:"created_at"`
	User            UserBrief `json:"user"`
}
type CommentForm struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	LikeCount int64     `json:"like_count"`
	CreatedAt time.Time `json:"created_at"`
	User      UserBrief `json:"user"`
}

// ==================== Auth ====================

// 通用用户信息结构体

// Register:
// RegisterParams 注册参数
type RegisterParams struct {
	StudentID string `json:"student_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Password  string `json:"password" binding:"required,min=6,max=20,alphanum"`
	Phone     string `json:"phone" binding:"required"`
	Email     string `json:"email" binding:"omitempty,email"`
	QQ        string `json:"qq" binding:"omitempty"`
	WeiXin    string `json:"weixin" binding:"omitempty"`
}

// Login:
// LoginParams 登录参数
type LoginParams struct {
	StudentID string `json:"student_id" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

// ChangePasswordParams 修改密码参数
type ChangePasswordParams struct {
	UserID      int64  `json:"-"`
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=20,alphanum"`
}

// UpdateProfileParams 修改用户信息参数
type UpdateProfileParams struct {
	UserID int64  `json:"-"`
	Name   string `json:"name" binding:"omitempty"`
	Phone  string `json:"phone" binding:"omitempty"`
	Email  string `json:"email" binding:"omitempty,email"`
	QQ     string `json:"qq" binding:"omitempty"`
	WeiXin string `json:"weixin" binding:"omitempty"`
}

type UpdateDescriptionParams struct {
	UserID      int64  `json:"-"`
	Description string `json:"description" binding:"required"`
}

// UserProfileResponse 用户首页信息（公开）
type UserProfileResponse struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	AvatarURL    string `json:"avatar_url"`
	Level        int    `json:"level"`
	Description  string `json:"description"`
	PrizeCount   int    `json:"prize_count"`
	PhotoCount   int64  `json:"photo_count"`
	AttemptCount int64  `json:"attempt_count"`
}

// UploadAvatarParams 上传头像参数
type UploadAvatarParams struct {
	UserID     int64                 `form:"-"`
	AvatarFile *multipart.FileHeader `form:"avatar" binding:"required"`
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

// CreatePhotoResponse 上传图片投稿响应
type CreatePhotoResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

// List:
// ListPhotoParams 图片列表查询参数
type ListPhotoParams struct {
	common.PagerForm
	Solved *bool  `form:"solved"`
	SortBy string `form:"sort_by" binding:"omitempty,oneof=created_at attempts_count likes_count"`
}

// ListPhotosResponse 图片列表响应
type ListPhotosResponse struct {
	Total  int64       `json:"total"`
	Photos []PhotoForm `json:"photos"`
}

// GetByID:
// GetPhotoParams 获取图片详情参数
type GetPhotoParams struct {
	PhotoID       int64 `uri:"id" binding:"min=1"`
	CurrentUserID int64 `json:"-"`
}

// WinnerInfo 获奖者信息

// PhotoDetailResponse 图片详情响应
type PhotoDetailResponse struct {
	ID            int64       `json:"id"`
	Title         string      `json:"title"`
	Description   string      `json:"description"`
	ImageURL      string      `json:"image_url"`
	Author        UserBrief   `json:"author"`
	Solved        bool        `json:"solved"`
	AttemptsCount int         `json:"attempts_count"`
	CreatedAt     time.Time   `json:"created_at"`
	Winner        AttemptForm `json:"winner,omitempty"`
}

// GetImageStream:
// ImageStream 封装图片流数据
type ImageStream struct {
	Reader      io.ReadCloser
	ContentType string
	Size        int64
	Filename    string
}

// ListUserPhotosParams 获取用户图片列表参数
type ListUserPhotosParams struct {
	common.PagerForm
	SortBy string `form:"sort_by" binding:"omitempty,oneof=created_at attempts_count likes_count"`
	UserID int64  `uri:"id" binding:"min=1"`
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

// AttemptShow:
// AttemptShowParams 获取答题记录参数
type AttemptShowParams struct {
	common.PagerForm
	// PhotoID int64 `uri:"id" binding:"min=1"`
	UserID int64  `uri:"user_id" binding:"min=1"`
	SortBy string `form:"sort_by" binding:"omitempty,oneof=created_at attempts_count likes_count"`
}

// AttemptShowResponse 答题记录响应
type AttemptShowResponse struct {
	Total    int64         `json:"total"`
	Attempts []AttemptForm `json:"attempts"`
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

// CreateCommentParams 创建评论参数
type CreateCommentParams struct {
	PhotoID     int64  `uri:"id" binding:"min=1"`
	UserID      int64  `form:"-"`
	CommentText string `json:"comment_text" binding:"required"`
}

// CreateCommentResponse 创建评论响应
type CreateCommentResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

// ListUserCommentsParams 获取用户评论列表参数
type ListUserCommentsParams struct {
	common.PagerForm
	SortBy string `form:"sort_by" binding:"omitempty,oneof=created_at,like_count"`
	UserID int64  `uri:"id" binding:"min=1"`
}

// ListUserCommentsResponse 用户评论列表响应
type ListUserCommentsResponse struct {
	Total    int64         `json:"total"`
	Comments []CommentForm `json:"comments"`
}

// ListUserAttemptsParams 获取用户答题列表参数
type ListUserAttemptsParams struct {
	common.PagerForm
	UserID int64 `uri:"id" binding:"min=1"`
}

// ListUserAttemptsResponse 用户答题列表响应
type ListUserAttemptsResponse struct {
	Total    int64         `json:"total"`
	Attempts []AttemptForm `json:"attempts"`
}

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
