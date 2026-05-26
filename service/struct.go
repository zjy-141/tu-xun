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
	LikesCount      int       `json:"likes_count"`
	CreatedAt       time.Time `json:"created_at"`
	User            UserBrief `json:"user"`
}
type CommentForm struct {
	ID         int64     `json:"id"`
	Content    string    `json:"content"`
	LikesCount int64     `json:"likes_count"`
	CreatedAt  time.Time `json:"created_at"`
	User       UserBrief `json:"user"`
}
type PrizeForm struct {
	ID         int64      `json:"id"`
	PhotoID    int64      `json:"photo_id"`
	PhotoTitle string     `json:"photo_title"`
	Status     string     `json:"status"`
	PrizeType  string     `json:"prize_type"`
	AwardedAt  *time.Time `json:"awarded_at"`
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
	PhotoID       int64
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
	UserID int64
}

// ==================== Attempt ====================

// Create:
// CreateAttemptParams 提交答题参数
type CreateAttemptParams struct {
	PhotoID         int64
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
	// PhotoID int64
	UserID int64
	SortBy string `form:"sort_by" binding:"omitempty,oneof=created_at attempts_count likes_count"`
}

// ==================== Admin ====================

// PendingPhotos:
// PendingPhotoForm 输入
type PendingPhotoParams struct {
	common.PagerForm
	Status     string `form:"status" binding:"omitempty,oneof=pending approved rejected"`
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
	PhotoID      int64
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
	Status     string `form:"status" binding:"omitempty,oneof=pending approved rejected"`
	AdminLevel int    //审核员等级
}

// PendingAttemptForm 待审核答题项
type PendingAttemptForm struct {
	AttemptID       int64     `json:"attempt_id"`
	PhotoID         int64     `json:"photo_id"`
	PhotoTitle      string    `json:"photo_title"`
	ImageURL        string    `json:"image_url"`        //猜测照片
	GuessedLocation string    `json:"guessed_location"` //猜测地址
	ThumbURL        string    `json:"thumb_url"`        //原照片
	LocationSecret  string    `json:"location_secret"`  //原照片的正确地址（仅管理员可见）
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
	AttemptID    int64
	Action       string `json:"action" binding:"required"`
	RejectReason string `json:"reject_reason"`
	Solved       string `json:"solved" binding:"omitempty,oneof=solved unsolved"` //管理员审核时是否标记图片为已破解（仅审核通过时有效）
	AdminLevel   int    //审核员等级
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
	CommentID    int64
	Action       string `json:"action" binding:"required"`
	RejectReason string `json:"reject_reason"`
}

// ReviewCommentResponse 审核评论响应
type ReviewCommentResponse struct {
	CommentID int64  `json:"comment_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// UpdateAdminLevelParams 高级管理员调整管理员等级参数
type UpdateAdminLevelParams struct {
	UserID        int64
	TargetLevel   int   `json:"target_level" binding:"required,min=0"`
	OperatorID    int64 `json:"-"` // 操作者 ID，由 controller 注入
	OperatorLevel int   `json:"-"` // 操作者等级，由 controller 注入
}

// UpdateAdminLevelResponse 调整管理员等级响应
type UpdateAdminLevelResponse struct {
	UserID   int64  `json:"user_id"`
	Name     string `json:"name"`
	OldLevel int    `json:"old_level"`
	NewLevel int    `json:"new_level"`
	Message  string `json:"message"`
}

// ==================== Prize ====================

// MyPrizes:
// PrizeItem 奖品项

type MyPrizesParams struct {
	common.PagerForm
	UserID int64 `json:"-" `
}

// MyPrizesResponse 我的奖品列表响应
type MyPrizesResponse struct {
	Total  int64       `json:"total"`
	Prizes []PrizeForm `json:"prizes"`
}

// ==================== Comment ====================

// CreateCommentParams 创建评论参数
type CreateCommentParams struct {
	PhotoID     int64
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
	SortBy string `form:"sort_by" binding:"omitempty,oneof=created_at likes_count"`
	UserID int64
}

// ListCommentsResponse 评论列表响应
type ListCommentsResponse struct {
	Total    int64         `json:"total"`
	Comments []CommentForm `json:"comments"`
}

// ListUserAttemptsParams 获取用户答题列表参数
type ListUserAttemptsParams struct {
	common.PagerForm
	UserID int64
}

// ListAttemptsResponse 答题列表响应
type ListAttemptsResponse struct {
	Total    int64         `json:"total"`
	Attempts []AttemptForm `json:"attempts"`
}

// ListPhotoCommentsParams 获取图片下评论列表参数
type ListPhotoCommentsParams struct {
	common.PagerForm
	SortBy  string `form:"sort_by" binding:"omitempty,oneof=created_at likes_count"`
	PhotoID int64
}

// ListPhotoAttemptsParams 获取图片下答题列表参数
type ListPhotoAttemptsParams struct {
	common.PagerForm
	SortBy  string `form:"sort_by" binding:"omitempty,oneof=created_at likes_count"`
	PhotoID int64
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

// ==================== Message ====================

// --- 通知消息（系统通知） ---

// ListMessageParams 消息列表查询参数
type ListMessageParams struct {
	common.PagerForm
	UserID int64 `json:"-"`
}

// MessageItem 消息项
type MessageItem struct {
	ID          int64     `json:"id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	RelatedID   int64     `json:"related_id"`
	RelatedType string    `json:"related_type"`
	IsRead      bool      `json:"is_read"`
	CreatedAt   time.Time `json:"created_at"`
}

// ListMessagesResponse 消息列表响应
type ListMessagesResponse struct {
	Total    int64         `json:"total"`
	Messages []MessageItem `json:"messages"`
}

// UnreadCountResponse 未读消息数响应
type UnreadCountResponse struct {
	Count int64 `json:"count"`
}

// --- 会话（微信风格聊天） ---

// ConversationItem 会话列表项（微信首页）
type ConversationItem struct {
	PartnerID     int64     `json:"partner_id"`
	PartnerName   string    `json:"partner_name"`
	PartnerAvatar string    `json:"partner_avatar"`
	LastContent   string    `json:"last_content"`
	LastTime      time.Time `json:"last_time"`
	UnreadCount   int64     `json:"unread_count"`
}

// ListConversationsResponse 会话列表响应
type ListConversationsResponse struct {
	Conversations []ConversationItem `json:"conversations"`
}

// ChatMessage 聊天消息（对话详情中的一条）
type ChatMessage struct {
	ID        int64     `json:"id"`
	SenderID  int64     `json:"sender_id"`
	Content   string    `json:"content"`
	IsMine    bool      `json:"is_mine"`
	Type      string    `json:"type"` // 消息类型(review_rejected/review_approved/system?/chat)
	CreatedAt time.Time `json:"created_at"`
}

// GetConversationParams 获取对话详情参数
type GetConversationParams struct {
	common.PagerForm
	UserID    int64 `json:"-"`
	PartnerID int64
}

// ConversationDetailResponse 对话详情响应
type ConversationDetailResponse struct {
	Partner  UserBrief     `json:"partner"`
	Messages []ChatMessage `json:"messages"`
	Total    int64         `json:"total"`
}

// SendChatParams 发送聊天消息参数
type SendChatParams struct {
	UserID    int64 `json:"-"`
	PartnerID int64
	Content   string `json:"content" binding:"required"`
}

// ==================== Like ====================

// ToggleLikeResponse 切换点赞响应
type ToggleLikeResponse struct {
	Liked bool  `json:"liked"`
	Count int64 `json:"count"`
}

// LikeStatusResponse 点赞状态响应
type LikeStatusResponse struct {
	Liked bool  `json:"liked"`
	Count int64 `json:"count"`
}
