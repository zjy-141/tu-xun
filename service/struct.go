package service

import (
	"mime/multipart"
	"time"
	"tu-xun/common"
)

// ==================== 公共 ====================

// ResponseIS 创建/更新操作的标准响应
type ResponseIS struct {
	ID         int64  `json:"id"`
	Status     string `json:"status"`
	VerifyCode string `json:"verify_code,omitempty"`
}

// StandardErrorResponse 公共错误响应 schema
type StandardErrorResponse struct {
	Success bool   `json:"success"`
	Resp    any    `json:"resp"`
	Message string `json:"message"`
	Code    uint64 `json:"code"`
}

// Location 地理位置坐标聚合对象
type Location struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	CoordType string  `json:"coord_type"`
}

// Media 标准媒体对象（图片/视频），origin_url 和 thumb_url 按接口场景下发
type Media struct {
	OriginURL string `json:"origin_url,omitempty"`
	ThumbURL  string `json:"thumb_url,omitempty"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}

// UserBrief 简要用户信息（嵌套用）
type UserBrief struct {
	ID       int64  `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// ActivityBrief 简要活动信息（嵌套用）
type ActivityBrief struct {
	ID        int64      `json:"id"`
	Title     string     `json:"title"`
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
}

// PhotoBrief 简要题目信息（用于答题记录等嵌套）
type PhotoBrief struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Image Media  `json:"image"`
}

// ==================== Test ====================

// TestLoginParams 测试登录参数
type TestLoginParams struct {
	UserID   int64  `form:"user_id" binding:"required"`
	Password string `form:"password" binding:"required"`
}

// ==================== User ====================

// LoginCallbackParams OAuth 回调参数
type LoginCallbackParams struct {
	Code string `form:"code" binding:"required"`
	// State       string `form:"state" binding:"required"`
	RedirectURI string `form:"redirect_uri" binding:"required,url"`
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

// UserSummary 用户信息响应
type UserSummary struct {
	ID                     int64  `json:"id"`
	NetID                  string `json:"netid"`
	Username               string `json:"username"`
	Nickname               string `json:"nickname"`
	Avatar                 string `json:"avatar"`
	Level                  int    `json:"level"`
	ScoreCount             int    `json:"score_count"`
	Status                 string `json:"status"`
	NicknameEditsRemaining int    `json:"nickname_edits_remaining"`
	AvatarEditsRemaining   int    `json:"avatar_edits_remaining"`
}

// LoginResult 登录成功响应（UserSummary + session_id）
type LoginResult struct {
	UserSummary
	SessionID string `json:"session_id"`
}

// UpdateNicknameParams 修改昵称参数
type UpdateNicknameParams struct {
	ID       int64  `json:"-"`
	Nickname string `json:"nickname" binding:"required,max=10"`
}

// UpdateNicknameResponse 修改昵称响应
type UpdateNicknameResponse struct {
	Nickname               string `json:"nickname"`
	NicknameEditsRemaining int    `json:"nickname_edits_remaining"`
}

// UploadAvatarParams 上传头像参数
type UploadAvatarParams struct {
	ID         int64                 `form:"-"`
	AvatarFile *multipart.FileHeader `form:"avatar" binding:"required"`
}

// UploadAvatarResponse 上传头像响应
type UploadAvatarResponse struct {
	Avatar               string `json:"avatar"`
	AvatarEditsRemaining int    `json:"avatar_edits_remaining"`
}

// ==================== Activity ====================

// ActivityCard 活动卡片（客户端与管理端共用）
type ActivityCard struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	CoverImage  Media     `json:"cover_image"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	PhotoCount  int       `json:"photo_count"`
}

// ActivityCardPage 活动分页
type ActivityCardPage struct {
	Total int64          `json:"total"`
	List  []ActivityCard `json:"list"`
}

// ActivityListParams 客户端活动列表参数
type ActivityListParams struct {
	common.PagerForm
	Status  string `form:"status" binding:"omitempty,oneof=active ended"`
	Keyword string `form:"keyword" binding:"omitempty,max=50"`
}

// ==================== Photo ====================

// PhotoCreateParams 投稿参数
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

// PhotoListParams 客户端题目列表参数
type PhotoListParams struct {
	common.PagerForm
	ActivityID int64  `form:"activity_id" binding:"omitempty"`
	Solved     *bool  `form:"solved" binding:"omitempty"`
	SortBy     string `form:"sort_by" binding:"omitempty,oneof=created_at hot"`
	Keyword    string `form:"keyword" binding:"omitempty,max=50"`
}

// PhotoCard 题目卡片（浏览列表用）
type PhotoCard struct {
	ID         int64         `json:"id"`
	Activity   ActivityBrief `json:"activity"`
	Author     UserBrief     `json:"author"`
	Title      string        `json:"title"`
	Image      Media         `json:"image"`
	LikesCount int           `json:"likes_count"`
	Liked      bool          `json:"liked"`
	Solved     bool          `json:"solved"`
	CreatedAt  *time.Time    `json:"created_at"`
}

// PhotoCardPage 题目卡片分页
type PhotoCardPage struct {
	Total int64       `json:"total"`
	List  []PhotoCard `json:"list"`
}

// PhotoDetail 题目详情
type PhotoDetail struct {
	ID                int64         `json:"id"`
	Activity          ActivityBrief `json:"activity"`
	Author            UserBrief     `json:"author"`
	Title             string        `json:"title"`
	Description       string        `json:"description"`
	Image             Media         `json:"image"`
	Location          *Location     `json:"location"`
	Solved            bool          `json:"solved"`
	SolvedCount       int           `json:"solved_count"`
	AttemptsCount     int           `json:"attempts_count"`
	UserAttemptsCount int           `json:"user_attempts_count"`
	LikesCount        int           `json:"likes_count"`
	Liked             bool          `json:"liked"`
	CreatedAt         *time.Time    `json:"created_at"`
	Status            string        `json:"status"`
}

// UserPhotoCard 我的投稿卡片
type UserPhotoCard struct {
	ID            int64         `json:"id"`
	Activity      ActivityBrief `json:"activity"`
	Title         string        `json:"title"`
	Image         Media         `json:"image"`
	AttemptsCount int           `json:"attempts_count"`
	SolvedCount   int           `json:"solved_count"`
	Status        string        `json:"status"`
	CreatedAt     *time.Time    `json:"created_at"`
}

// UserPhotoCardPage 我的投稿分页
type UserPhotoCardPage struct {
	Total int64           `json:"total"`
	List  []UserPhotoCard `json:"list"`
}

// PhotosListUserParams 我的投稿记录参数
type PhotosListUserParams struct {
	common.PagerForm
	UserID     int64  `form:"-"`
	ActivityID int64  `form:"activity_id" binding:"omitempty"`
	Status     string `form:"status" binding:"omitempty,oneof=pending approved rejected"`
}

// UserPhotoDetail 我的投稿详情（仅 pending/rejected）
type UserPhotoDetail struct {
	ID           int64         `json:"id"`
	Activity     ActivityBrief `json:"activity"`
	Title        string        `json:"title"`
	Description  string        `json:"description"`
	Image        Media         `json:"image"`
	Location     Location      `json:"location"`
	Status       string        `json:"status"`
	RejectReason string        `json:"reject_reason"`
	CreatedAt    *time.Time    `json:"created_at"`
}

// ==================== Attempt ====================

// AttemptCreateParams 提交作答参数
type AttemptCreateParams struct {
	UserID    int64                 `form:"-"`
	PhotoID   int64                 `form:"-"`
	ImageFile *multipart.FileHeader `form:"image_file" binding:"required"`
	Longitude float64               `form:"longitude" binding:"required"`
	Latitude  float64               `form:"latitude" binding:"required"`
	CoordType string                `form:"coord_type" binding:"required,oneof=wgs84 gcj02 bd09"`
}

// AttemptRecord 作答记录基底（在本图的作答列表用）
type AttemptRecord struct {
	ID           int64      `json:"id"`
	Image        Media      `json:"image"`
	Location     Location   `json:"location"`
	Status       string     `json:"status"`
	RejectReason string     `json:"reject_reason,omitempty"`
	CreatedAt    *time.Time `json:"created_at"`
}

// AttemptRecordPage 作答记录分页
type AttemptRecordPage struct {
	Total int64           `json:"total"`
	List  []AttemptRecord `json:"list"`
}

// UserAttemptCard 我的作答记录卡片
type UserAttemptCard struct {
	ID                int64      `json:"id"`
	Photo             PhotoBrief `json:"photo"`
	UserAttemptsCount int        `json:"user_attempts_count"`
	Status            string     `json:"status"`
	CreatedAt         *time.Time `json:"created_at"`
}

// UserAttemptCardPage 我的作答记录分页
type UserAttemptCardPage struct {
	Total int64             `json:"total"`
	List  []UserAttemptCard `json:"list"`
}

// AttemptsListUserParams 我的作答记录参数
type AttemptsListUserParams struct {
	common.PagerForm
	UserID     int64  `form:"-"`
	ActivityID int64  `form:"activity_id" binding:"omitempty"`
	Status     string `form:"status" binding:"omitempty,oneof=pending unsolved solved"`
}

// PhotoAttemptsUserListParams 某题目下的当前用户作答记录参数
type PhotoAttemptsUserListParams struct {
	common.PagerForm
	UserID  int64 `form:"-"`
	PhotoID int64 `form:"-"`
}

// SolvesListParams 破解记录列表参数（公开，仅 solved）
type SolvesListParams struct {
	common.PagerForm
	PhotoID int64 `form:"-"`
}

// SolveItem 破解记录列表项
type SolveItem struct {
	ID         int64      `json:"id"`
	Author     UserBrief  `json:"author"`
	Image      Media      `json:"image"`
	Location   Location   `json:"location"`
	LikesCount int        `json:"likes_count"`
	Liked      bool       `json:"liked"`
	CreatedAt  *time.Time `json:"created_at"`
}

// SolveItemPage 破解记录分页
type SolveItemPage struct {
	Total int64       `json:"total"`
	List  []SolveItem `json:"list"`
}

// ==================== Comment ====================

// CommentCreateParams 发表评论参数
type CommentCreateParams struct {
	UserID      int64  `json:"-"`
	PhotoID     int64  `json:"-"`
	CommentText string `json:"content" binding:"required,min=1,max=140"`
}

// CommentListParams 评论列表参数
type CommentListParams struct {
	common.PagerForm
	PhotoID int64  `form:"-"`
	SortBy  string `form:"sort_by" binding:"omitempty,oneof=created_at likes_count"`
}

// CommentItem 评论列表项
type CommentItem struct {
	ID         int64      `json:"id"`
	Author     UserBrief  `json:"author"`
	Content    string     `json:"content"`
	LikesCount int        `json:"likes_count"`
	Liked      bool       `json:"liked"`
	CreatedAt  *time.Time `json:"created_at"`
}

// CommentItemPage 评论分页
type CommentItemPage struct {
	Total int64         `json:"total"`
	List  []CommentItem `json:"list"`
}

// CommentDeleteParams 删除评论参数
type CommentDeleteParams struct {
	UserID    int64 `json:"-"`
	Level     int   `json:"-"`
	CommentID int64 `uri:"id" binding:"required"`
}

// ==================== Like ====================

// LikeResult 点赞操作结果
type LikeResult struct {
	Liked      bool  `json:"liked"`
	LikesCount int64 `json:"likes_count"`
}

// ==================== Score ====================

// ScoreLogParams 积分流水查询参数
type ScoreLogParams struct {
	UserID int64 `form:"-"`
	common.PagerForm
}

// ScoreLogItem 积分流水记录
type ScoreLogItem struct {
	ID           int64      `json:"id"`
	Delta        int        `json:"delta"`
	Balance      int        `json:"balance"`
	Reason       string     `json:"reason"`
	RelatedID    int64      `json:"related_id"`
	RelatedType  string     `json:"related_type"`
	RelatedTitle *string    `json:"related_title"`
	CreatedAt    *time.Time `json:"created_at"`
}

// ScoreLogPage 积分流水分页
type ScoreLogPage struct {
	Total        int64          `json:"total"`
	List         []ScoreLogItem `json:"list"`
	TotalIncome  int            `json:"total_income"`
	TotalExpense int            `json:"total_expense"`
}

// ScoreChangeParams 积分变更参数
type ScoreChangeParams struct {
	UserID      int64
	Delta       int
	Reason      string
	RelatedID   int64
	RelatedType string
	Remark      string
}

// ==================== Good ====================

// GoodListParams 客户端奖品列表参数
type GoodListParams struct {
	common.PagerForm
	Keyword string `form:"keyword" binding:"omitempty,max=50"`
}

// GoodItem 奖品列表项（客户端与管理端共用）
type GoodItem struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Image       Media      `json:"image"`
	ScorePrice  int        `json:"score_price"`
	Stock       int        `json:"stock"`
	Status      string     `json:"status"`
	CreatedAt   *time.Time `json:"created_at"`
}

// GoodItemPage 奖品分页
type GoodItemPage struct {
	Total int64      `json:"total"`
	List  []GoodItem `json:"list"`
}

// GoodBrief 兑换记录中的奖品摘要
type GoodBrief struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Image      Media  `json:"image"`
	ScorePrice int    `json:"score_price"`
}

// ==================== Exchange ====================

// ExchangeCreateParams 兑换奖品参数
type ExchangeCreateParams struct {
	GoodID         int64  `json:"good_id"`
	UserID         int64  `json:"-"`
	Quantity       int    `json:"quantity" binding:"required,min=1"`
	IdempotencyKey string `json:"-"`
}

// ExchangeListParams 兑换记录列表参数
type ExchangeListParams struct {
	common.PagerForm
	UserID int64  `form:"-"`
	Status string `form:"status" binding:"omitempty,oneof=pending verified cancelled"`
}

// ExchangeItem 兑换记录列表项
type ExchangeItem struct {
	ID         int64      `json:"id"`
	Good       GoodBrief  `json:"good"`
	Quantity   int        `json:"quantity"`
	ScoreCost  int        `json:"score_cost"`
	Status     string     `json:"status"`
	VerifyCode string     `json:"verify_code"`
	ExchangeAt *time.Time `json:"exchange_at"`
	CreatedAt  *time.Time `json:"created_at"`
}

// ExchangeItemPage 兑换记录分页
type ExchangeItemPage struct {
	Total int64          `json:"total"`
	List  []ExchangeItem `json:"list"`
}

// ==================== Announcement（通知/公告） ====================

// AnnouncementListParams 客户端通知列表参数
type AnnouncementListParams struct {
	common.PagerForm
	UserID  int64  `form:"-"`
	Keyword string `form:"keyword" binding:"omitempty,max=50"`
}

// AnnouncementListItem 通知列表项
type AnnouncementListItem struct {
	ID             int64      `json:"id"`
	Title          string     `json:"title"`
	ContentPreview string     `json:"content_preview"`
	IsRead         bool       `json:"is_read"`
	CreatedAt      *time.Time `json:"created_at"`
}

// AnnouncementPage 通知分页（客户端，含未读数）
type AnnouncementPage struct {
	Total       int64                  `json:"total"`
	List        []AnnouncementListItem `json:"list"`
	UnreadCount int64                  `json:"unread_count"`
}

// AnnouncementDetail 通知详情
type AnnouncementDetail struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Image       *Media     `json:"image,omitempty"`
	RelatedType string     `json:"related_type,omitempty"`
	RelatedID   int64      `json:"related_id,omitempty"`
	IsRead      bool       `json:"is_read"`
	CreatedAt   *time.Time `json:"created_at"`
}

// CreateAnnouncementRequest 发布通知请求（管理员，multipart）
type CreateAnnouncementRequest struct {
	Title       string                `form:"title" binding:"required,max=20"`
	Content     string                `form:"content" binding:"required,max=10000"`
	ImageFile   *multipart.FileHeader `form:"image_file" binding:"omitempty"`
	RelatedType string                `form:"related_type" binding:"omitempty,oneof=activity"`
	RelatedID   int64                 `form:"related_id" binding:"omitempty"`
}

// UpdateAnnouncementRequest 更新通知请求（管理员，multipart）
type UpdateAnnouncementRequest struct {
	ID             int64                 `form:"-"`
	Title          string                `form:"title" binding:"omitempty,max=20"`
	Content        string                `form:"content" binding:"omitempty,max=10000"`
	ImageFile      *multipart.FileHeader `form:"image_file" binding:"omitempty"`
	RemoveImage    bool                  `form:"remove_image" binding:"omitempty"`
	RemoveRelation bool                  `form:"remove_relation" binding:"omitempty"`
	RelatedType    string                `form:"related_type" binding:"omitempty,oneof=activity"`
	RelatedID      int64                 `form:"related_id" binding:"omitempty"`
}

// AdminAnnouncementListParams 管理端通知列表参数
type AdminAnnouncementListParams struct {
	common.PagerForm
	Keyword string `form:"keyword" binding:"omitempty,max=50"`
}

// AdminAnnouncementListItem 管理端通知列表项（含已读人数）
type AdminAnnouncementListItem struct {
	ID             int64      `json:"id"`
	Title          string     `json:"title"`
	ContentPreview string     `json:"content_preview"`
	ReadCount      int64      `json:"read_count"`
	CreatedAt      *time.Time `json:"created_at"`
}

// AdminAnnouncementPage 管理端通知分页
type AdminAnnouncementPage struct {
	Total int64                       `json:"total"`
	List  []AdminAnnouncementListItem `json:"list"`
}

// AdminAnnouncementDetail 管理端通知详情（无 is_read）
type AdminAnnouncementDetail struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Image       *Media     `json:"image,omitempty"`
	RelatedType string     `json:"related_type,omitempty"`
	RelatedID   int64      `json:"related_id,omitempty"`
	ReadCount   int64      `json:"read_count"`
	CreatedAt   *time.Time `json:"created_at"`
}

// ==================== InteractionMessage（互动消息） ====================

// InteractionMessageListParams 互动消息列表参数
type InteractionMessageListParams struct {
	common.PagerForm
	UserID int64  `form:"-"`
	Type   string `form:"type" binding:"omitempty,oneof=like comment"`
}

// InteractionMessage 互动消息
type InteractionMessage struct {
	ID          int64      `json:"id"`
	Type        string     `json:"type"`
	User        UserBrief  `json:"user"`
	RelatedType string     `json:"related_type"`
	RelatedID   int64      `json:"related_id"`
	PhotoID     int64      `json:"photo_id"`
	Content     string     `json:"content"`
	IsRead      bool       `json:"is_read"`
	CreatedAt   *time.Time `json:"created_at"`
}

// InteractionMessagePage 互动消息分页
type InteractionMessagePage struct {
	Total       int64                `json:"total"`
	List        []InteractionMessage `json:"list"`
	UnreadCount int64                `json:"unread_count"`
}

// ==================== ContentBlock（内容位） ====================

// ContentBlock 内容位
type ContentBlock struct {
	Key       string     `json:"key"`
	Content   string     `json:"content"`
	RelatedID int64      `json:"related_id"`
	Version   int        `json:"version"`
	UpdatedAt *time.Time `json:"updated_at"`
}

// UpdateContentRequest 更新内容位请求
type UpdateContentRequest struct {
	Content   string `json:"content" binding:"required,max=10000"`
	RelatedID int64  `json:"related_id" binding:"omitempty"`
}

// ==================== Feedback ====================

// FeedbackCreateParams 提交反馈参数
type FeedbackCreateParams struct {
	UserID    int64                 `form:"-"`
	Title     string                `form:"title" binding:"required,max=20"`
	Content   string                `form:"content" binding:"required,min=1,max=500"`
	Type      int                   `form:"type" binding:"required,oneof=1 2 3 4"`
	Phone     string                `form:"phone" binding:"omitempty,max=20"`
	MediaFile *multipart.FileHeader `form:"media_file" binding:"omitempty"`
}

// FeedbackListParams 反馈列表查询参数
type FeedbackListParams struct {
	common.PagerForm
	Type        int    `form:"type" binding:"omitempty,oneof=1 2 3 4"`
	Status      string `form:"status" binding:"omitempty,oneof=pending resolved"`
	Keyword     string `form:"keyword" binding:"omitempty,max=50"`
	UserKeyword string `form:"user_keyword" binding:"omitempty,max=50"`
}

// FeedbackItem 反馈列表项
type FeedbackItem struct {
	ID        int64      `json:"id"`
	User      UserBrief  `json:"user"`
	Title     string     `json:"title"`
	Type      int        `json:"type"`
	Status    string     `json:"status"`
	CreatedAt *time.Time `json:"created_at"`
}

// FeedbackPage 反馈列表分页
type FeedbackPage struct {
	Total int64          `json:"total"`
	List  []FeedbackItem `json:"list"`
}

// FeedbackMedia 反馈附件
type FeedbackMedia struct {
	ID        int64  `json:"id"`
	OriginURL string `json:"origin_url"`
	ThumbURL  string `json:"thumb_url,omitempty"`
	Width     *int   `json:"width"`
	Height    *int   `json:"height"`
	MediaType int    `json:"media_type"`
}

// FeedbackDetail 反馈详情
type FeedbackDetail struct {
	ID        int64           `json:"id"`
	User      UserBrief       `json:"user"`
	Title     string          `json:"title"`
	Content   string          `json:"content"`
	Type      int             `json:"type"`
	Phone     string          `json:"phone"`
	Status    string          `json:"status"`
	Medias    []FeedbackMedia `json:"medias"`
	CreatedAt *time.Time      `json:"created_at"`
}

// FeedbackReviewParams 处理反馈参数
type FeedbackReviewParams struct {
	FeedbackID int64  `json:"-"`
	Status     string `json:"status" binding:"required,oneof=pending resolved"`
}

// ==================== Admin Photo ====================

// AdminPhotoListParams 管理端题目列表参数
type AdminPhotoListParams struct {
	common.PagerForm
	Status      string  `form:"status" binding:"omitempty,oneof=pending approved rejected"`
	ActivityIDs []int64 `form:"activity_ids" binding:"omitempty"`
	Solved      *bool   `form:"solved" binding:"omitempty"`
	Keyword     string  `form:"keyword" binding:"omitempty,max=50"`
	UserKeyword string  `form:"user_keyword" binding:"omitempty,max=50"`
}

// AdminPhotoListItem 管理端题目列表项
type AdminPhotoListItem struct {
	ID            int64         `json:"id"`
	Activity      ActivityBrief `json:"activity"`
	Author        UserBrief     `json:"author"`
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	Image         Media         `json:"image"`
	Location      Location      `json:"location"`
	AttemptsCount int           `json:"attempts_count"`
	SolvedCount   int           `json:"solved_count"`
	LikesCount    int           `json:"likes_count"`
	Status        string        `json:"status"`
	RejectReason  string        `json:"reject_reason,omitempty"`
	CreatedAt     *time.Time    `json:"created_at"`
}

// AdminPhotoListPage 管理端题目列表分页
type AdminPhotoListPage struct {
	Total int64                `json:"total"`
	List  []AdminPhotoListItem `json:"list"`
}

// AdminPhotoCreateForm 管理员新增题目的 multipart 表单
type AdminPhotoCreateForm struct {
	ActivityID  int64                 `form:"activity_id" binding:"required"`
	Title       string                `form:"title" binding:"required,max=20"`
	Description string                `form:"description" binding:"omitempty,max=50"`
	ImageFile   *multipart.FileHeader `form:"image_file" binding:"required"`
	Longitude   float64               `form:"longitude" binding:"required"`
	Latitude    float64               `form:"latitude" binding:"required"`
	CoordType   string                `form:"coord_type" binding:"required,oneof=wgs84 gcj02 bd09"`
}

// AdminPhotoUpdateForm 管理员更新题目的 multipart 表单
type AdminPhotoUpdateForm struct {
	PhotoID     int64                 `form:"-"`
	Title       string                `form:"title" binding:"omitempty,max=20"`
	Description string                `form:"description" binding:"omitempty,max=50"`
	ImageFile   *multipart.FileHeader `form:"image_file" binding:"omitempty"`
	Longitude   float64               `form:"longitude" binding:"omitempty"`
	Latitude    float64               `form:"latitude" binding:"omitempty"`
	CoordType   string                `form:"coord_type" binding:"omitempty,oneof=wgs84 gcj02 bd09"`
}

// AdminReviewPhotoParams 审核题目参数
type AdminReviewPhotoParams struct {
	PhotoID      int64
	Action       string `json:"action" binding:"required,oneof=approve reject"`
	RejectReason string `json:"reject_reason" binding:"omitempty,max=50"`
	AdminLevel   int
}

// ==================== Admin Attempt ====================

// AdminAttemptListParams 管理端作答列表参数
type AdminAttemptListParams struct {
	common.PagerForm
	Status       string `form:"status" binding:"omitempty,oneof=pending solved unsolved"`
	Keyword      string `form:"keyword" binding:"omitempty,max=50"`
	PhotoKeyword string `form:"photo_keyword" binding:"omitempty,max=50"`
	UserKeyword  string `form:"user_keyword" binding:"omitempty,max=50"`
}

// AdminAttemptPhotoBrief 管理端作答列表中的题目摘要（含坐标）
type AdminAttemptPhotoBrief struct {
	ID       int64    `json:"id"`
	Title    string   `json:"title"`
	Image    Media    `json:"image"`
	Location Location `json:"location"`
}

// AdminAttemptListItem 管理端作答列表项
type AdminAttemptListItem struct {
	ID            int64                  `json:"id"`
	User          UserBrief              `json:"user"`
	Photo         AdminAttemptPhotoBrief `json:"photo"`
	GuessImage    Media                  `json:"guess_image"`
	GuessLocation Location               `json:"guess_location"`
	Status        string                 `json:"status"`
	RejectReason  string                 `json:"reject_reason,omitempty"`
	CreatedAt     *time.Time             `json:"created_at"`
}

// AdminAttemptListPage 管理端作答列表分页
type AdminAttemptListPage struct {
	Total int64                  `json:"total"`
	List  []AdminAttemptListItem `json:"list"`
}

// AdminReviewAttemptParams 审核作答参数
type AdminReviewAttemptParams struct {
	AttemptID    int64
	Solved       string `json:"solved" binding:"required,oneof=solved unsolved"`
	RejectReason string `json:"reject_reason" binding:"omitempty,max=50"`
	AdminLevel   int
}

// ==================== Admin Comment ====================

// AdminCommentListParams 管理端评论列表参数
type AdminCommentListParams struct {
	common.PagerForm
	Status       string `form:"status" binding:"omitempty,oneof=pending approved rejected"`
	Keyword      string `form:"keyword" binding:"omitempty,max=50"`
	PhotoKeyword string `form:"photo_keyword" binding:"omitempty,max=50"`
	UserKeyword  string `form:"user_keyword" binding:"omitempty,max=50"`
}

// AdminCommentPhotoBrief 管理端评论列表中的题目摘要
type AdminCommentPhotoBrief struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

// AdminCommentListItem 管理端评论列表项
type AdminCommentListItem struct {
	ID        int64                  `json:"id"`
	User      UserBrief              `json:"user"`
	Photo     AdminCommentPhotoBrief `json:"photo"`
	Content   string                 `json:"content"`
	Status    string                 `json:"status"`
	CreatedAt *time.Time             `json:"created_at"`
}

// AdminCommentListPage 管理端评论列表分页
type AdminCommentListPage struct {
	Total int64                  `json:"total"`
	List  []AdminCommentListItem `json:"list"`
}

// AdminReviewCommentParams 审核评论参数
type AdminReviewCommentParams struct {
	CommentID    int64
	Action       string `json:"action" binding:"required,oneof=approve reject"`
	RejectReason string `json:"reject_reason" binding:"omitempty,max=50"`
}

// ==================== Admin Activity ====================

// AdminActivityListParams 管理端活动列表参数
type AdminActivityListParams struct {
	common.PagerForm
	Keyword string `form:"keyword" binding:"omitempty,max=50"`
	Status  string `form:"status" binding:"omitempty,oneof=not_started active ended"`
}

// AdminActivityCreate 创建活动参数
type AdminActivityCreate struct {
	Title       string                `form:"title" binding:"required,max=20"`
	CoverFile   *multipart.FileHeader `form:"cover_file" binding:"required"`
	Description string                `form:"description" binding:"required,max=100"`
	StartTime   *time.Time            `form:"start_time" binding:"required"`
	EndTime     *time.Time            `form:"end_time" binding:"required"`
	RewardTiers string                `form:"reward_tiers" binding:"omitempty"` // JSON: [{"batch":1,"rank_limit":3,"attempt_points":20},...]
}

// AdminActivityUpdate 更新活动参数
type AdminActivityUpdate struct {
	ActivityID  int64                 `form:"-"`
	Title       string                `form:"title" binding:"omitempty,max=20"`
	CoverFile   *multipart.FileHeader `form:"cover_file" binding:"omitempty"`
	Description string                `form:"description" binding:"omitempty,max=100"`
	StartTime   *time.Time            `form:"start_time" binding:"omitempty"`
	EndTime     *time.Time            `form:"end_time" binding:"omitempty"`
	RewardTiers string                `form:"reward_tiers" binding:"omitempty"` // JSON: [{"batch":1,"rank_limit":3,"attempt_points":20},...]
}

// ==================== Admin Good ====================

// AdminGoodListParams 管理端奖品列表参数
type AdminGoodListParams struct {
	common.PagerForm
	Status  string `form:"status" binding:"omitempty,oneof=in_store out_store"`
	Keyword string `form:"keyword" binding:"omitempty,max=50"`
}

// GoodCreateParams 创建奖品参数
type GoodCreateParams struct {
	Name        string                `form:"name" binding:"required,max=20"`
	Description string                `form:"description" binding:"omitempty,max=50"`
	NeedScore   int                   `form:"score_price" binding:"required,min=0"`
	Stock       int                   `form:"stock" binding:"required,min=0"`
	ImageFile   *multipart.FileHeader `form:"image_file" binding:"required"`
	Status      string                `form:"status" binding:"omitempty,oneof=in_store out_store"`
}

// GoodUpdateParams 更新奖品参数
type GoodUpdateParams struct {
	GoodID      int64                 `form:"-"`
	Name        string                `form:"name" binding:"omitempty,max=20"`
	Description string                `form:"description" binding:"omitempty,max=50"`
	NeedScore   int                   `form:"score_price" binding:"omitempty,min=0"`
	Stock       int                   `form:"stock" binding:"omitempty,min=0"`
	ImageFile   *multipart.FileHeader `form:"image_file" binding:"omitempty"`
	Status      string                `form:"status" binding:"omitempty,oneof=in_store out_store"`
}

// ==================== Admin Exchange ====================

// AdminExchangeListParams 管理端兑换列表参数
type AdminExchangeListParams struct {
	common.PagerForm
	Status      string `form:"status" binding:"omitempty,oneof=pending verified cancelled"`
	Keyword     string `form:"keyword" binding:"omitempty,max=50"`
	VerifyCode  string `form:"verify_code" binding:"omitempty"`
	UserKeyword string `form:"user_keyword" binding:"omitempty,max=50"`
	GoodKeyword string `form:"good_keyword" binding:"omitempty,max=50"`
}

// AdminExchangeItem 管理端兑换记录列表项
type AdminExchangeItem struct {
	ID         int64      `json:"id"`
	User       UserBrief  `json:"user"`
	Good       GoodBrief  `json:"good"`
	Quantity   int        `json:"quantity"`
	ScoreCost  int        `json:"score_cost"`
	Status     string     `json:"status"`
	VerifyCode string     `json:"verify_code"`
	ExchangeAt *time.Time `json:"exchange_at"`
	CreatedAt  *time.Time `json:"created_at"`
}

// AdminExchangePage 管理端兑换记录分页
type AdminExchangePage struct {
	Total int64               `json:"total"`
	List  []AdminExchangeItem `json:"list"`
}

// AdminExchangeVerifyParams 核销/取消兑换参数
type AdminExchangeVerifyParams struct {
	ExchangeID int64  `json:"-"`
	Action     string `json:"action" binding:"required,oneof=verify cancel"`
}

// ==================== Admin Stats ====================

// AdminStats 管理端工作台统计
type AdminStats struct {
	UserCount            int64 `json:"user_count"`
	PendingPhotoCount    int64 `json:"pending_photo_count"`
	PendingAttemptCount  int64 `json:"pending_attempt_count"`
	PendingCommentCount  int64 `json:"pending_comment_count"`
	PendingFeedbackCount int64 `json:"pending_feedback_count"`
}

// ==================== Admin User ====================

// AdminUserListParams 管理端用户列表参数（Level 3 专用）
type AdminUserListParams struct {
	common.PagerForm
	Keyword string `form:"keyword" binding:"omitempty,max=50"`
	Status  string `form:"status" binding:"omitempty,oneof=active banned"`
	Level   int    `form:"level" binding:"omitempty,oneof=1 2 3"`
}

// AdminUserPage 管理端用户列表分页
type AdminUserPage struct {
	Total int64         `json:"total"`
	List  []UserSummary `json:"list"`
}

// AdminUpdateLevelParams 调整权限等级参数
type AdminUpdateLevelParams struct {
	UserID        int64 `json:"-"`
	TargetLevel   int   `json:"target_level" binding:"required,oneof=1 2"`
	OperatorID    int64 `json:"-"`
	OperatorLevel int   `json:"-"`
}

// AdminSetUserStatusParams 封禁/解封用户参数
type AdminSetUserStatusParams struct {
	UserID        int64  `json:"-"`
	Status        string `json:"status" binding:"required,oneof=banned active"`
	OperatorID    int64  `json:"-"`
	OperatorLevel int    `json:"-"`
}
