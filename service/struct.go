package service

import (
	"io"
	"mime/multipart"
	"time"
	"tu-xun/common"
)

// ==================== 公共 ====================

type Redirect struct {
	Redirect_url string `json:"redirect_url" form:"redirect_url" uri:"redirect_url"`
}

type ResponseIS struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

// 简要用户信息
type UserBrief struct {
	ID        int64  `json:"id"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

// 简要活动信息
type ActivityBrief struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

// ==================== Test ====================

// TestLoginParams 测试登录参数
type TestLoginParams struct {
	NetID    string `form:"netid"`
	Username string `form:"username"`
	Password string `form:"password" binding:"required"`
}

// ==================== User ====================

type LoginCallbackParams struct {
	Code  string `form:"code" binding:"required"`
	State string `form:"state" binding:"required"`
}

// StudentOauthInfo tz统一认证返回的用户信息
type StudentOauthInfo struct {
	Name    string   `json:"name"`
	Netid   string   `json:"netid"`
	Role    string   `json:"role"`
	Roles   []string `json:"roles"`
	Scope   []string `json:"scope"`
	Service string   `json:"service"`
	Sub     string   `json:"sub"`
}

// 返回值示例
//
//	{
//	  "email": "--",
//	  "name": "张继尧",
//	  "netid": "2251416412",
//	  "role": "user",
//	  "roles": [
//	    "user"
//	  ],
//	  "scope": [
//	    "openid profile"
//	  ],
//	  "service": "demo",
//	  "sub": "2251416412"
//	}
//
// LoginCallback 返回值（UserSummary）
type UserForm struct {
	ID        int64  `json:"id"`
	NetID     string `json:"netid"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
	Level     int    `json:"level"`
	Status    string `json:"status"`
}

// 更新昵称
type UserUpdateParams struct {
	ID       int64  `json:"-"`
	Nickname string `json:"nickname" binding:"optional,max=10"`
}

// 更新头像
type UserUploadAvatar struct {
	ID         int64                 `form:"-"`
	AvatarFile *multipart.FileHeader `form:"avatar" binding:"required"`
}

// ==================== Activity ====================

// ActivityForm 活动信息（公开列表用 ActivityCard）
type ActivityForm struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	CoverURL    string    `json:"cover_url"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
}

// ActivityForms 活动信息列表
type ActivityForms struct {
	Total      int64          `json:"total"`
	List []ActivityForm `json:"list"`
}

// ==================== Photo ====================

// PhotoCreateParams 上传图片参数
type PhotoCreateParams struct {
	UserID      int64                 `form:"-"`
	ActivityID  int64                 `form:"activity_id" binding:"required"`
	Title       string                `form:"title" binding:"required,max=20"`
	Description string                `form:"description" binding:"omitempty,max=50"`
	ImageFile   *multipart.FileHeader `form:"image_file" binding:"required"`
	Longitude   float64               `form:"longitude" binding:"required"`
	Latitude    float64               `form:"latitude" binding:"required"`
	CoordType   string                `form:"coord_type" binding:"required,oneof=wgs84 gcj02 bd09"`
}

// PhotoListParams 图片列表参数
type PhotoListParams struct {
	common.PagerForm
	ActivityID int64  `form:"activity_id" binding:"omitempty"`
	Solved     *bool  `form:"solved" binding:"omitempty"`
	SortBy     string `form:"sort_by" binding:"omitempty,oneof=created_at likes_count attempts_count"`
	Keyword    string `form:"keyword" binding:"omitempty,max=50"`
}

// PhotoForm 图片信息
type PhotoForm struct {
	ID          int64         `json:"id"`
	Author      UserBrief     `json:"author"`
	Activity    ActivityBrief `json:"activity,omitempty"`
	Title       string        `json:"title"`
	Description string        `json:"description,omitempty"`
	ThumbURL    string        `json:"thumb_url"`
	Solved      bool          `json:"solved"`
	LikesCount  int           `json:"likes_count"`
	CreatedAt   *time.Time    `json:"created_at"`
}

// PhotoForms 图片信息列表

type PhotoForms struct {
	Total  int64       `json:"total"`
	List []PhotoForm `json:"list"`
}

// PhotoGetByIDParams 获取图片详情参数
type PhotoGetByIDParams struct {
	PhotoID int64 `json:"-"`
}

// PhotoDetail 图片详情
type PhotoDetail struct {
	ID            int64         `json:"id"`
	Author        UserBrief     `json:"author"`
	Activity      ActivityBrief `json:"activity"`
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	ImageURL      string        `json:"image_url"`
	Longitude     float64       `json:"longitude,omitempty"`
	Latitude      float64       `json:"latitude,omitempty"`
	CoordType     string        `json:"coord_type,omitempty"`
	Solved        bool          `json:"solved"`
	AttemptsCount int           `json:"attempts_count"`
	LikesCount    int           `json:"likes_count"`
	CreatedAt     *time.Time    `json:"created_at"`
	Status        string        `json:"status"`
}

// ImageStream 图片流
type ImageStream struct {
	Reader      io.ReadCloser
	ContentType string
	Filename    string
	Size        int64
}

// PhotoAttemptsListParams 获取图片答题记录列表参数
type PhotoAttemptsListParams struct {
	common.PagerForm
	PhotoID int64  `form:"-"`
	SortBy  string `form:"sort_by" binding:"omitempty,oneof=created_at likes_count"`
}

// PhotoCommentsListParams 获取图片评论列表参数
type PhotoCommentsListParams struct {
	common.PagerForm
	PhotoID int64  `form:"-"`
	SortBy  string `form:"sort_by" binding:"omitempty,oneof=created_at likes_count"`
}

// PhotoAttemptsUserListParams 获取图片答题记录列表参数
type PhotoAttemptsUserListParams struct {
	common.PagerForm
	UserID  int64  `form:"-"`
	PhotoID int64  `form:"-"`
	SortBy  string `form:"sort_by" binding:"omitempty,oneof=created_at likes_count"`
	Status  string `form:"status" binding:"omitempty,oneof=pending unsolved solved"`
}

// PhotosListUserParams 获取该用户投稿的图片列表
type PhotosListUserParams struct {
	common.PagerForm
	UserID     int64  `form:"-"`
	ActivityID int64  `form:"activity_id" binding:"omitempty"`
	Solved     *bool  `form:"solved" binding:"omitempty,oneof=pending approved rejected"`
	SortBy     string `form:"sort_by" binding:"omitempty,oneof=created_at likes_count attempts_count"`
}

// UserPhotoForm 图片信息
type UserPhotoForm struct {
	ID int64 `json:"id"`
	// Author       UserBrief `json:"author"`
	Activity     ActivityBrief `json:"activity"`
	Title        string        `json:"title"`
	Description  string        `json:"description"`
	ThumbURL     string        `json:"thumb_url"`
	Solved       bool          `json:"solved"`
	LikesCount   int           `json:"likes_count"`
	CreatedAt    *time.Time    `json:"created_at"`
	Status       string        `json:"status"`
	RejectReason string        `json:"reject_reason"`
}

// UserPhotoForms 图片信息列表

type UserPhotoForms struct {
	Total  int64           `json:"total"`
	List []UserPhotoForm `json:"list"`
}

// PhotoDetailUserParams 获取图片详情参数
type PhotoDetailUserParams struct {
	PhotoID int64 `json:"-"`
	UserID  int64 `json:"-"`
	Level   int   `json:"-"`
}

type UserPhotoDetail struct {
	ID int64 `json:"id"`
	// Author       UserBrief `json:"author"`
	Activity      ActivityBrief `json:"activity"`
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	ImageURL      string        `json:"image_url"`
	Longitude     float64       `json:"longitude"`
	Latitude      float64       `json:"latitude"`
	CoordType     string        `json:"coord_type"`
	Solved        bool          `json:"solved"`
	LikesCount    int           `json:"likes_count"`
	AttemptsCount int           `json:"attempts_count"`
	CreatedAt     *time.Time    `json:"created_at"`
	Status        string        `json:"status"`
	RejectReason  string        `json:"reject_reason"`
}

// ==================== Attempt ====================

// AttemptCreateParams 提交答题参数
type AttemptCreateParams struct {
	UserID    int64                 `form:"-"`
	PhotoID   int64                 `form:"-"`
	ImageFile *multipart.FileHeader `form:"image_file" binding:"required"`
	Longitude float64               `form:"longitude" binding:"required"`
	Latitude  float64               `form:"latitude" binding:"required"`
	CoordType string                `form:"coord_type" binding:"required,oneof=wgs84 gcj02 bd09"`
}

// PhotoBrief 图片简要信息（用于答题列表项嵌套）
type PhotoBrief struct {
	ID       int64         `json:"id"`
	Title    string        `json:"title"`
	ThumbURL string        `json:"thumb_url"`
	Activity ActivityBrief `json:"activity"`
}

// AttemptForm 答题信息
type AttemptForm struct {
	ID         int64      `json:"id"`
	Author     UserBrief  `json:"author"`
	Photo      PhotoBrief `json:"photo"`
	ImageURL   string     `json:"image_url"`
	Solved     bool       `json:"solved"`
	LikesCount int        `json:"likes_count"`
	CreatedAt  *time.Time `json:"created_at"`
}

// AttemptForms 答题信息列表
type AttemptForms struct {
	Total    int64         `json:"total"`
	List []AttemptForm `json:"list"`
}

// AttemptsListUserParams 获取该用户的答题列表
type AttemptsListUserParams struct {
	common.PagerForm
	UserID     int64  `form:"-"`
	ActivityID int64  `form:"activity_id" binding:"omitempty"`
	Status     string `form:"status" binding:"omitempty,oneof=pending unsolved solved"`
	SortBy     string `form:"sort_by" binding:"omitempty,oneof=created_at likes_count"`
}

// UserAttemptForm 答题信息
type UserAttemptForm struct {
	ID int64 `json:"id"`
	// Author       UserBrief `json:"author"`
	Photo        PhotoBrief `json:"photo"`
	ImageURL     string     `json:"image_url"`
	Longitude    float64    `json:"longitude"`
	Latitude     float64    `json:"latitude"`
	CoordType    string     `json:"coord_type"`
	LikesCount   int        `json:"likes_count"`
	CreatedAt    *time.Time `json:"created_at"`
	Status       string     `json:"status"`
	RejectReason string     `json:"reject_reason,omitempty"`
}

// UserAttemptForms 答题信息列表
type UserAttemptForms struct {
	Total    int64             `json:"total"`
	List []UserAttemptForm `json:"list"`
}

// ==================== Comment ====================

// CommentCreateParams 提交评论参数
type CommentCreateParams struct {
	UserID      int64  `json:"-"`
	PhotoID     int64  `json:"-"`
	CommentText string `json:"comment_text" binding:"required,max=140"`
}

// CommentForm 评论信息
type CommentForm struct {
	ID          int64      `json:"id"`
	Author      UserBrief  `json:"author"`
	PhotoID     int64      `json:"photo_id"`
	CommentText string     `json:"comment_text"`
	LikesCount  int        `json:"likes_count"`
	CreatedAt   *time.Time `json:"created_at"`
}

// CommentForms 评论信息列表
type CommentForms struct {
	Total    int64         `json:"total"`
	List []CommentForm `json:"list"`
}

// CommentDeleteParams 删除评论参数
type CommentDeleteParams struct {
	UserID    int64 `json:"-"`
	Level     int   `json:"-"`
	CommentID int64 `uri:"comment_id" binding:"required"`
}

// ==================== Like ====================

type LikeTarget struct {
	UserID     int64  `json:"-"`
	TargetType string `json:"-"`
	TargetID   int64  `json:"-"`
	IsLike     *bool  `json:"is_like" binding:"required"`
}

type LikeCount struct {
	IsLike     bool  `json:"is_like"`
	LikesCount int64 `json:"likes_count"`
}

// ==================== Score ====================
// /  Score 当前积分
type ScoreTotal struct {
	TotalScore int `json:"total_score"`
}

type ScoreLogParams struct {
	UserID int64 `form:"-"`
	common.PagerForm
}

// ScoreForm 积分信息
type ScoreLogForm struct {
	ID          int64      `json:"id"`
	Delta       int        `json:"delta"`
	Balance     int        `json:"balance"`
	Reason      string     `json:"reason"`
	RelatedID   int64      `json:"related_id"`
	RelatedType string     `json:"related_type"`
	CreatedAt   *time.Time `json:"created_at"`
}

// ScoreForms 积分信息列表
type ScoreLogForms struct {
	Total     int64          `json:"total"`
	List []ScoreLogForm `json:"list"`
}

// RegularScoreReward 常规积分奖励参数
type ScoreChangeParams struct {
	UserID      int64
	Delta       int
	Reason      string
	RelatedID   int64
	RelatedType string
	Remark      string
}

// ==================== Good ====================
// PhotoListParams 图片列表参数
type GoodListParams struct {
	common.PagerForm
	Available bool `form:"available" binding:"omitempty"`
}

// GoodForm 奖品信息
type GoodForm struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	ThumbURL  string `json:"thumb_url"`
	NeedScore int    `json:"need_score"`
	Stock     int    `json:"stock"`
}

// GoodForms 奖品信息列表
type GoodForms struct {
	Total int64      `json:"total"`
	List []GoodForm `json:"list"`
}

// GoodGetByIDParams 获取奖品详情参数
type GoodGetByIDParams struct {
	GoodID int64 `json:"-"`
}

type GoodDetail struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	NeedScore   int    `json:"need_score"`
	Stock       int    `json:"stock"`
}

// ==================== Exchange ====================

// 兑换奖品信息
type ExchangeClaim struct {
	GoodID         int64  `json:"good_id"`
	UserID         int64  `json:"-"`
	Quantity       int    `json:"quantity"`
	IdempotencyKey string `json:"-"` // 从请求头 Idempotency-Key 注入
}

// PhotoListParams 图片列表参数
type ExchangeListParams struct {
	common.PagerForm
	UserID int64  `form:"-"`
	Status string `form:"status" binding:"omitempty,oneof=pending verified cancelled"`
}

// ExchangeForm 兑换信息
type ExchangeForm struct {
	ID         int64      `json:"id"`
	Good       GoodForm   `json:"good"`
	Quantity   int        `json:"quantity"`
	ScoreCost  int        `json:"score_cost"`
	Status     string     `json:"status"`
	ExchangeAt *time.Time `json:"exchange_at"`
	CreatedAt  *time.Time `json:"created_at"`
}

// ExchangeForms 兑换信息列表
type ExchangeForms struct {
	Total     int64          `json:"total"`
	List []ExchangeForm `json:"list"`
}

// ==================== Notification (统一通知，原 Message + Notice 合并) ====================

// NotificationListParams 通知列表查询参数
type NotificationListParams struct {
	common.PagerForm
	UserID      int64  `form:"-"`
	Category    string `form:"category" binding:"omitempty,oneof=normal interaction"`
	Type        string `form:"type" binding:"omitempty,oneof=general global_announcement like comment review"`
	RelatedType string `form:"related_type" binding:"omitempty,oneof=activity photo attempt comment feedback"`
	RelatedID   int64  `form:"related_id" binding:"omitempty"`
	Keyword     string `form:"keyword" binding:"omitempty,max=50"`
}

// NotificationListItem 通知列表项（只含摘要字段）
type NotificationListItem struct {
	ID             int64      `json:"id"`
	Category       string     `json:"category"`
	Type           string     `json:"type"`
	Title          string     `json:"title"`
	ContentPreview string     `json:"content_preview"`
	IsRead         bool       `json:"is_read"`
	CreatedAt      *time.Time `json:"created_at"`
}

// NotificationForms 通知列表
type NotificationForms struct {
	Total int64                  `json:"total"`
	List  []NotificationListItem `json:"list"`
}

// NotificationGetByIDParams 获取通知详情参数
type NotificationGetByIDParams struct {
	NotificationID int64 `json:"-"`
}

// NotificationDetail 通知详情
type NotificationDetail struct {
	ID          int64      `json:"id"`
	SenderID    int64      `json:"sender_id,omitempty"`
	Category    string     `json:"category"`
	Type        string     `json:"type"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	ImageURL    string     `json:"image_url,omitempty"`
	RelatedID   int64      `json:"related_id,omitempty"`
	RelatedType string     `json:"related_type,omitempty"`
	IsRead      bool       `json:"is_read"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   *time.Time `json:"created_at"`
}

// NotificationReadParams 标记通知为已读
type NotificationReadParams struct {
	UserID         int64 `json:"-"`
	NotificationID int64 `json:"-"`
}

// NotificationUnreadCount 未读通知总数
type NotificationUnreadCount struct {
	Count int64 `json:"count"`
}

// CreateNotificationRequest 创建通知请求（管理员，multipart）
type CreateNotificationRequest struct {
	Type        string                `form:"type" binding:"required,oneof=general global_announcement"`
	Title       string                `form:"title" binding:"required,max=20"`
	Content     string                `form:"content" binding:"required,max=2000"`
	ImageFile   *multipart.FileHeader `form:"image_file" binding:"omitempty"`
	RelatedType string                `form:"related_type" binding:"omitempty,oneof=activity"`
	RelatedID   int64                 `form:"related_id" binding:"omitempty"`
	ExpiresAt   *time.Time            `form:"expires_at" binding:"omitempty"`
}

// UpdateNotificationRequest 更新通知请求（管理员，multipart）
type UpdateNotificationRequest struct {
	ID             int64                 `form:"-"`
	Type           string                `form:"type" binding:"omitempty,oneof=general global_announcement"`
	Title          string                `form:"title" binding:"omitempty,max=20"`
	Content        string                `form:"content" binding:"omitempty,max=2000"`
	ImageFile      *multipart.FileHeader `form:"image_file" binding:"omitempty"`
	RemoveImage    bool                  `form:"remove_image" binding:"omitempty"`
	RemoveRelation bool                  `form:"remove_relation" binding:"omitempty"`
	RelatedType    string                `form:"related_type" binding:"omitempty,oneof=activity"`
	RelatedID      int64                 `form:"related_id" binding:"omitempty"`
	ExpiresAt      *time.Time            `form:"expires_at" binding:"omitempty"`
}

// ==================== Feedback ====================

// FeedbackCreateParams 发送反馈
type FeedbackCreateParams struct {
	UserID     int64                 `form:"-"`
	Title      string                `form:"title" binding:"required,max=20"`
	Content    string                `form:"content" binding:"required,max=500"`
	Type       int                   `form:"type" binding:"required,oneof=1 2 3 4"`
	Phone      string                `form:"phone" binding:"omitempty,max=20"`
	ImageFile1 *multipart.FileHeader `form:"image_file1" binding:"omitempty"`
	ImageFile2 *multipart.FileHeader `form:"image_file2" binding:"omitempty"`
	ImageFile3 *multipart.FileHeader `form:"image_file3" binding:"omitempty"`
}

// FeedbackListParams 反馈列表查询参数
type FeedbackListParams struct {
	common.PagerForm
	// UserID int64 `form:"-"`
	Type    int    `form:"type" binding:"omitempty,oneof=1 2 3 4"` // 1内容 2玩法 3技术 4其他
	Status  string `form:"status" binding:"omitempty,oneof=pending resolved"`
	Keyword string `form:"keyword" binding:"omitempty,max=50"`
}

// FeedbackForm 反馈列表项
type FeedbackForm struct {
	ID        int64      `json:"id"`
	Title     string     `json:"title"`
	Type      int        `json:"type"`
	Status    string     `json:"status"`
	CreatedAt *time.Time `json:"created_at"`
}

// FeedbackForms 反馈列表
type FeedbackForms struct {
	Total     int64          `json:"total"`
	List []FeedbackForm `json:"list"`
}

// FeedbackGetByIDParams 获取反馈详情参数
type FeedbackGetByIDParams struct {
	FeedbackID int64 `json:"-"`
}

// FeedbackDetail 反馈详情
type FeedbackDetail struct {
	ID        int64               `json:"id"`
	UserID    int64               `json:"user_id"`
	Title     string              `json:"title"`
	Content   string              `json:"content"`
	Type      int                 `json:"type"`
	Phone     string              `json:"phone"`
	Status    string              `json:"status"`
	Medias    []FeedbackMediaForm `json:"medias"`
	CreatedAt *time.Time          `json:"created_at"`
}

// FeedbackMediaForm 反馈附件
type FeedbackMediaForm struct {
	ID        int64  `json:"id"`
	URL       string `json:"url"`
	MediaType int    `json:"media_type"`
}

// FeedbackGetByIDParams 获取反馈详情参数
type FeedbackReviewParams struct {
	FeedbackID int64  `json:"-"`
	Status     string `json:"status" binding:"required,oneof=pending resolved"`
}

// ==================== Admin ====================

// AdminPhotoListParams 管理端统一题目池查询参数
type AdminPhotoListParams struct {
	common.PagerForm
	Status      string `form:"status" binding:"omitempty,oneof=pending approved rejected"`
	ActivityIDs []int64 `form:"activity_ids" binding:"omitempty"` // 重复参数，OR 匹配
	Keyword     string `form:"keyword" binding:"omitempty,max=50"`
}

// AdminPhotoListItem 管理端题目池列表项
type AdminPhotoListItem struct {
	ID            int64         `json:"id"`
	Activity      ActivityBrief `json:"activity"`
	Author        UserBrief     `json:"author"`
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	ImageURL      string        `json:"image_url"`
	ThumbURL      string        `json:"thumb_url"`
	Longitude     float64       `json:"longitude"`
	Latitude      float64       `json:"latitude"`
	CoordType     string        `json:"coord_type"`
	Solved        bool          `json:"solved"`
	AttemptsCount int           `json:"attempts_count"`
	LikesCount    int           `json:"likes_count"`
	Status        string        `json:"status"`
	RejectReason  string        `json:"reject_reason,omitempty"`
	CreatedAt     *time.Time    `json:"created_at"`
}

type AdminPhotoListForms struct {
	Total int64                `json:"total"`
	List  []AdminPhotoListItem `json:"list"`
}

// AdminPhotoUpsertForm 管理员新增/编辑活动题目的 multipart 表单
type AdminPhotoUpsertForm struct {
	Title       string                `form:"title" binding:"omitempty,max=20"`
	Description string                `form:"description" binding:"omitempty,max=50"`
	ImageFile   *multipart.FileHeader `form:"image_file" binding:"omitempty"`
	Longitude   float64               `form:"longitude" binding:"omitempty"`
	Latitude    float64               `form:"latitude" binding:"omitempty"`
	CoordType   string                `form:"coord_type" binding:"omitempty,oneof=wgs84 gcj02 bd09"`
}

// ReviewPhotoParams 审核图片参数
type AdminReviewPhotoParams struct {
	PhotoID      int64
	Action       string `json:"action" binding:"required,oneof=approve reject"`
	RejectReason string `json:"reject_reason" binding:"omitempty,max=50"`
	AdminLevel   int    //审核员等级
}

// AdminAttemptListParams 管理端答题列表查询参数
type AdminAttemptListParams struct {
	common.PagerForm
	Status  string `form:"status" binding:"omitempty,oneof=pending solved unsolved"`
	Keyword string `form:"keyword" binding:"omitempty,max=50"`
}

// AdminAttemptListItem 管理端答题列表项
type AdminAttemptListItem struct {
	AttemptID      int64      `json:"attempt_id"`
	PhotoID        int64      `json:"photo_id"`
	PhotoTitle     string     `json:"photo_title"`
	GuessImageURL  string     `json:"guess_image_url"`
	GuessLongitude float64    `json:"guess_longitude"`
	GuessLatitude  float64    `json:"guess_latitude"`
	GuessCoordType string     `json:"guess_coord_type"`
	ThumbURL       string     `json:"thumb_url"`
	Longitude      float64    `json:"longitude"`
	Latitude       float64    `json:"latitude"`
	CoordType      string     `json:"coord_type"`
	Status         string     `json:"status"`
	SubmittedAt    *time.Time `json:"submitted_at"`
}

// AdminAttemptListForms 管理端答题列表响应
type AdminAttemptListForms struct {
	Total int64                 `json:"total"`
	List  []AdminAttemptListItem `json:"list"`
}

// AdminReviewAttemptParams 审核答题参数
type AdminReviewAttemptParams struct {
	AttemptID int64
	Solved    string `json:"solved" binding:"required,oneof=solved unsolved"` //管理员审核时是否标记图片为已破解（仅审核通过时有效）
	// Action       string `json:"action" binding:"required,oneof=approve reject"`
	RejectReason string `json:"reject_reason"  binding:"omitempty,max=50"`
	AdminLevel   int    //审核员等级
}

// AdminCommentListParams 管理端评论列表查询参数
type AdminCommentListParams struct {
	common.PagerForm
	Status  string `form:"status" binding:"omitempty,oneof=pending approved rejected"`
	Keyword string `form:"keyword" binding:"omitempty,max=50"`
}

// AdminCommentListItem 管理端评论列表项
type AdminCommentListItem struct {
	CommentID  int64      `json:"comment_id"`
	PhotoID    int64      `json:"photo_id"`
	PhotoTitle string     `json:"photo_title"`
	User       UserBrief  `json:"user"`
	Comment    string     `json:"comment"`
	Status     string     `json:"status"`
	CreatedAt  *time.Time `json:"created_at"`
}

// AdminCommentListForms 管理端评论列表响应
type AdminCommentListForms struct {
	Total int64                  `json:"total"`
	List  []AdminCommentListItem `json:"list"`
}

// AdminReviewComment:
// AdminReviewCommentParams 审核评论参数
type AdminReviewCommentParams struct {
	CommentID    int64
	Action       string `json:"action" binding:"required,oneof=approve reject"`
	RejectReason string `json:"reject_reason" binding:"omitempty,max=50"`
}

// AdminUserListParams 创世管理员精确筛选参数
type AdminUserListParams struct {
	common.PagerForm
	NetID    string `form:"netid" binding:"omitempty,max=50"`
	Name     string `form:"name" binding:"omitempty,max=50"`
	Nickname string `form:"nickname" binding:"omitempty,max=50"`
	Status   string `form:"status" binding:"omitempty,oneof=active banned"`
	Level    int    `form:"level" binding:"omitempty,oneof=1 2 3"`
}

// AdminSearchUsersParams 管理员搜索用户
type AdminSearchUsersParams struct {
	common.PagerForm
	Keyword string `form:"keyword" binding:"omitempty,max=50"` // 空或省略时不筛选，返回全部用户
	Status  string `form:"status" binding:"omitempty,oneof=active banned"`
	Level   int    `form:"level" binding:"omitempty,oneof=1 2 3"`
}

// AdminUserForms 用户列表响应
type AdminUserForms struct {
	Total int64      `json:"total"`
	List []UserForm `json:"list"`
}

// UpdateAdminLevelParams 高级管理员调整管理员等级参数
type AdminUpdateLevelParams struct {
	ID            int64 `json:"id" binding:"required"`
	TargetLevel   int   `json:"target_level" binding:"required,oneof=1 2"`
	OperatorID    int64 `json:"-"` // 操作者 ID，由 controller 注入
	OperatorLevel int   `json:"-"` // 操作者等级，由 controller 注入
}

// AdminSetUserStatusParams 封禁/解封用户参数
type AdminSetUserStatusParams struct {
	UserID       int64  `json:"-"`
	Status       string `json:"status" binding:"required,oneof=banned active"`
	OperatorID   int64  `json:"-"` // 操作者 ID
	OperatorLevel int   `json:"-"` // 操作者等级
}

// ==================== AdminGood ====================

type AdminListGoodsParams struct {
	common.PagerForm
	Available bool   `form:"available" binding:"omitempty"`
	Status    string `form:"status" binding:"omitempty,oneof=inStore outStore"`
	Keyword   string `form:"keyword" binding:"omitempty,max=50"`
}

// GoodForm 奖品信息
type AdminGoodForm struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ThumbURL    string     `json:"thumb_url"`
	NeedScore   int        `json:"need_score"`
	Stock       int        `json:"stock"`
	Status      string     `json:"status"`
	CreatedAt   *time.Time `json:"created_at"`
}

// GoodForms 奖品信息列表
type AdminGoodForms struct {
	Total int64           `json:"total"`
	List []AdminGoodForm `json:"list"`
}

// GoodGetByIDParams 获取奖品详情参数
type AdminGoodGetByIDParams struct {
	GoodID int64 `json:"-"`
}

type AdminGoodDetail struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ImageURL    string     `json:"image_url"`
	NeedScore   int        `json:"need_score"`
	Stock       int        `json:"stock"`
	Status      string     `json:"status"`
	CreatedAt   *time.Time `json:"created_at"`
}

// GoodCreateParams 创建商品参数
type GoodCreateParams struct {
	Name        string                `form:"name" binding:"required,max=20"`
	Description string                `form:"description" binding:"omitempty,max=50"`
	NeedScore   int                   `form:"need_score" binding:"required,min=0"`
	Stock       int                   `form:"stock" binding:"required,min=0"`
	ImageFile   *multipart.FileHeader `form:"image" binding:"required"`
	Status      string                `form:"status" binding:"omitempty,oneof=inStore outStore"`
}

// GoodUpdateParams 更新商品参数
type GoodUpdateParams struct {
	GoodID      int64                 `form:"-"`
	Name        string                `form:"name" binding:"omitempty,max=20"`
	Description string                `form:"description" binding:"omitempty,max=50"`
	NeedScore   int                   `form:"need_score" binding:"omitempty,min=0"`
	Stock       int                   `form:"stock" binding:"omitempty,min=0"`
	ImageFile   *multipart.FileHeader `form:"image" binding:"omitempty"`
	Status      string                `form:"status" binding:"omitempty,oneof=inStore outStore"`
}

// GoodGetByIDParams 获取奖品详情参数
type AdminGoodStatusParams struct {
	GoodID int64  `json:"-"`
	Status string `json:"status" binding:"required,oneof=inStore outStore"`
}

// GoodUpdateStockParams 更新商品库存参数

type GoodUpdateStockParams struct {
	GoodID int64 `json:"-"`
	Stock  int   `json:"stock" binding:"required,min=0"`
}

type GoodStock struct {
	ID    int64 `json:"id"`
	Stock int   `json:"stock"`
}

// ==================== AdminExchange ====================

type AdminExchangeListParams struct {
	common.PagerForm
	Status       string `form:"status" binding:"omitempty,oneof=pending verified cancelled"`
	ExchangeID   int64  `form:"exchange_id" binding:"omitempty"`
	UserKeyword  string `form:"user_keyword" binding:"omitempty,max=50"`
	GoodKeyword  string `form:"good_keyword" binding:"omitempty,max=50"`
}
type AdminExchangeVerifyParams struct {
	ExchangeID int64  `json:"exchange_id" binding:"required"`
	Action     string `json:"action" binding:"required,oneof=verify cancel"`
}

// ExchangeForm 兑换信息
type AdminExchangeForm struct {
	ID         int64      `json:"id"`
	User       UserBrief  `json:"user"`
	Good       GoodForm   `json:"good"`
	Quantity   int        `json:"quantity"`
	ScoreCost  int        `json:"score_cost"`
	Status     string     `json:"status"`
	ExchangeAt *time.Time `json:"exchange_at"`
	CreatedAt  *time.Time `json:"created_at"`
}

// ExchangeForms 兑换信息列表
type AdminExchangeForms struct {
	Total          int64               `json:"total"`
	List []AdminExchangeForm `json:"list"`
}

// ==================== AdminActivity ====================

// AdminActivityListParams 活动列表参数
type AdminActivityListParams struct {
	common.PagerForm
	Keyword string `form:"keyword" binding:"omitempty,max=50"`
	Status  string `form:"status" binding:"omitempty,oneof=not_started active ended"`
}

// RewardTierInput 奖励阶梯入参
type RewardTierInput struct {
	Batch         int `json:"batch" binding:"required,min=1"`
	RankLimit     int `json:"rank_limit" binding:"required,min=1"`
	AttemptPoints int `json:"attempt_points" binding:"required,min=0"`
}

// AdminActivityForm 活动信息
type AdminActivityForm struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	CoverURL    string    `json:"cover_url"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
}

// AdminActivityForms 活动信息列表
type AdminActivityForms struct {
	Total      int64               `json:"total"`
	List []AdminActivityForm `json:"list"`
}

// PhotoGetByIDParams 获取图片详情参数
type AdminActivityGetByIDParams struct {
	ActivityID int64 `json:"-"`
}

// AdminActivityDetail 活动详情
type AdminActivityDetail struct {
	ID           int64             `json:"id"`
	Title        string            `json:"title"`
	CoverURL     string            `json:"cover_url"`
	Description  string            `json:"description"`
	StartTime    time.Time         `json:"start_time"`
	EndTime      time.Time         `json:"end_time"`
	PhotoPoints  int               `json:"photo_points"`
	RewardTiers  []RewardTierInput `json:"reward_tiers"`
}

// AdminActivityCreate 创建活动参数
type AdminActivityCreate struct {
	Title       string                `form:"title" binding:"required,max=20"`
	CoverFile   *multipart.FileHeader `form:"cover_file" binding:"required"`
	Description string                `form:"description" binding:"required,max=100"`
	StartTime   *time.Time            `form:"start_time" binding:"required"`
	EndTime     *time.Time            `form:"end_time" binding:"required"`
	PhotoPoints *int                  `form:"photo_points" binding:"required"`
	RewardTiers string                `form:"reward_tiers" binding:"omitempty"` // 改为 string
}

// AdminActivityUpdate 更新活动参数
type AdminActivityUpdate struct {
	ActivityID  int64                 `form:"activity_id" binding:"required"`
	Title       string                `form:"title" binding:"omitempty,max=20"`
	CoverFile   *multipart.FileHeader `form:"cover_file" binding:"required"`
	Description string                `form:"description" binding:"omitempty,max=100"`
	StartTime   *time.Time            `form:"start_time" binding:"omitempty"`
	EndTime     *time.Time            `form:"end_time" binding:"omitempty"`
	PhotoPoints *int                  `form:"photo_points" binding:"omitempty,min=0"`
	RewardTiers string                `form:"reward_tiers" binding:"omitempty"` // JSON string
}
