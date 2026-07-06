package service

import (
	"io"
	"mime/multipart"
	"tu-xun/common"
)

// ==================== 公共 ====================

type Redirect struct {
	Redirect_url string `json:"redirect_url" form:"redirect_url" uri:"redirect_url"`
}

type Guid struct {
	Guid string `json:"guid" form:"guid" uri:"guid" binding:"required"`
}

type ResponseIM struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

// 简要用户信息
type UserBrief struct {
	ID        int64  `json:"id"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

// ==================== User ====================

// StudentOauthInfo 学生统一认证返回的用户信息
type StudentOauthInfo struct {
	MemberId   string `json:"memberId"`
	OrgId      string `json:"ordId"`
	OrgName    string `json:"orgName"`
	MemberName string `json:"memberName"`
	MemberCode string `json:"memberCode"`
	Email      string `json:"email"`
	Mobile     string `json:"mobile"`
	Photo      string `json:"photo"`
	ExtId      string `json:"extId"`
	Remark     string `json:"remark"`
	Tmail      string `json:"tmail"`
	CryptoMail string `json:"cryptoMail"`
	Sex        string `json:"sex"`
	Netid      string `json:"netid"`
	UserType   int    `json:"userType"`
	DeptInfos  []struct {
		Status       int    `json:"status"`
		DepId        string `json:"deptId"`
		DeptName     string `json:"deptName"`
		DeptCode     string `json:"deptCode"`
		DeptNode     string `json:"deptNode"`
		PositionInfo string `json:"positionInfo"`
		Employeeno   string `json:"employeeno"`
	} `json:"deptInfos"`
	UserTypes []struct {
		UserType     int    `json:"userType"`
		MemberNumber string `json:"memberNumber"`
		MemberName   string `json:"memberName"`
		Userid       string `json:"userid"`
	}
	LoginUserType int    `json:"loginUserType"`
	LoginPersonNo string `json:"loginPersonNo"`
}

// LoginCallback 返回值
type UserForm struct {
	ID        int64  `json:"id"`
	NetID     string `json:"netid"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
	Level     int    `json:"level"`
}

// 更新昵称
type UserUpdateParams struct {
	ID       int64  `json:"-"`
	Nickname string `json:"nickname" binding:"optional,max=20"`
}

// 更新头像
type UserUploadAvatar struct {
	ID         int64                 `form:"-"`
	AvatarFile *multipart.FileHeader `form:"avatar" binding:"required"`
}

// ==================== Activity ====================

// ActivityForm 活动信息
type ActivityForm struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Cover       string `json:"cover"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
}

// ActivityForms 活动信息列表
type ActivityForms struct {
	Total      int64          `json:"total"`
	Activities []ActivityForm `json:"activities"`
}

// ==================== Photo ====================

// PhotoCreateParams 上传图片参数
type PhotoCreateParams struct {
	UserID      int64                 `json:"-"`
	ActivityID  int64                 `json:"activity_id" binding:"required"`
	Title       string                `json:"title" binding:"required"`
	Description string                `json:"description"`
	ImageFile   *multipart.FileHeader `json:"image" binding:"required"`
	Longitude   float64               `json:"longitude" binding:"required"`
	Latitude    float64               `json:"latitude" binding:"required"`
}

// PhotoListParams 图片列表参数
type PhotoListParams struct {
	common.PagerForm
	ActivityID int64  `form:"activity_id" binding:"required"`
	Solved     *bool  `form:"solved" binding:"omitempty"`
	SortBy     string `form:"sort_by" binding:"omitempty,oneof=created_at likes_count attempts_count"`
	Keyword    string `form:"keyword" binding:"omitempty,max=50"`
}

// PhotoForm 图片信息
type PhotoForm struct {
	ID         int64     `json:"id"`
	Author     UserBrief `json:"author"`
	Title      string    `json:"title"`
	ThumbURL   string    `json:"thumb_url"`
	Solved     bool      `json:"solved"`
	LikesCount int       `json:"likes_count"`
}

// PhotoForms 图片信息列表

type PhotoForms struct {
	Total  int64       `json:"total"`
	Photos []PhotoForm `json:"photos"`
}

// PhotoGetByIDParams 获取图片详情参数
type PhotoGetByIDParams struct {
	PhotoID int64 `json:"-"`
}

// PhotoDetail 图片详情
type PhotoDetail struct {
	ID            int64     `json:"id"`
	Author        UserBrief `json:"author"`
	ActivityID    int64     `json:"activity_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	ImageURL      string    `json:"image_url"`
	Solved        bool      `json:"solved"`
	AttemptsCount int       `json:"attempts_count"`
	LikesCount    int       `json:"likes_count"`
	CreatedAt     string    `json:"created_at"`
	Status        string    `json:"status"`
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
	SortBy  string `form:"sort_by" binding:"omitempty,oneof=created_at likes_count attempts_count"`
}

// PhotoCommentsListParams 获取图片评论列表参数
type PhotoCommentsListParams struct {
	common.PagerForm
	PhotoID int64  `form:"-"`
	SortBy  string `form:"sort_by" binding:"omitempty,oneof=created_at likes_count attempts_count"`
}

// ==================== Attempt ====================

// AttemptCreateParams 提交答题参数
type AttemptCreateParams struct {
	UserID      int64                 `json:"-"`
	PhotoID     int64                 `json:"-"`
	CommentText string                `json:"comment_text" binding:"omitempty,max=500"`
	ImageFile   *multipart.FileHeader `json:"image_file" binding:"required"`
	Longitude   float64               `json:"longitude" binding:"required"`
	Latitude    float64               `json:"latitude" binding:"required"`
}

// AttemptForm 答题信息
type AttemptForm struct {
	ID          int64     `json:"id"`
	Author      UserBrief `json:"author"`
	PhotoID     int64     `json:"photo_id"`
	CommentText string    `json:"comment_text"`
	ImageURL    string    `json:"image_url"`
	Solved      bool      `json:"solved"`
	LikesCount  int       `json:"likes_count"`
	CreatedAt   string    `json:"created_at"`
}

// AttemptForms 答题信息列表
type AttemptForms struct {
	Total    int64         `json:"total"`
	Attempts []AttemptForm `json:"attempts"`
}

// ==================== Comment ====================

// CommentCreateParams 提交评论参数
type CommentCreateParams struct {
	UserID      int64  `json:"-"`
	PhotoID     int64  `json:"-"`
	CommentText string `json:"comment_text" binding:"omitempty,max=500"`
}

// CommentForm 评论信息
type CommentForm struct {
	ID          int64     `json:"id"`
	Author      UserBrief `json:"author"`
	PhotoID     int64     `json:"photo_id"`
	CommentText string    `json:"comment_text"`
	LikesCount  int       `json:"likes_count"`
	CreatedAt   string    `json:"created_at"`
}

// CommentForms 评论信息列表
type CommentForms struct {
	Total    int64         `json:"total"`
	Comments []CommentForm `json:"comments"`
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
	TargetType string `json:"target_type"`
	TargetID   int64  `json:"-"`
}

type LikeResponse struct {
	Liked     bool  `json:"is_like"`
	LikeCount int64 `json:"like_count"`
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
	ID          int64  `json:"id"`
	Delta       int    `json:"delta"`
	Balance     int    `json:"balance"`
	Reason      string `json:"reason"`
	RelatedID   int64  `json:"related_id"`
	RelatedType string `json:"related_type"`
	CreatedAt   string `json:"created_at"`
}

// ScoreForms 积分信息列表
type ScoreLogForms struct {
	Total     int64          `json:"total"`
	ScoreLogs []ScoreLogForm `json:"score_logs"`
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
	Goods []GoodForm `json:"goods"`
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
type ExchangeClain struct {
	GoodID   int64 `json:"good_id"`
	UserID   int64 `json:"-"`
	Quantity int   `json:"quantity"`
}

// PhotoListParams 图片列表参数
type ExchangeListParams struct {
	common.PagerForm
	UserID int64  `form:"-"`
	Status string `form:"status" binding:"omitempty,oneof=pending verified cancelled"`
}

// ExchangeForm 兑换信息
type ExchangeForm struct {
	ID         int64    `json:"id"`
	Good       GoodForm `json:"good"`
	Quantity   int      `json:"quantity"`
	ScoreCost  int      `json:"score_cost"`
	Status     string   `json:"status"`
	ExchangeAt string   `json:"exchange_at"`
	CreatedAt  string   `json:"created_at"`
}

// ExchangeForms 兑换信息列表
type ExchangeForms struct {
	Total     int64          `json:"total"`
	Exchanges []ExchangeForm `json:"exchanges"`
}

// ==================== Message ====================

// MessageListParams 消息列表查询参数
type MessageListParams struct {
	common.PagerForm
	UserID int64 `form:"-"`
}

// MessageForm 消息信息
type MessageForm struct {
	ID        int64  `json:"id"`
	SenderID  int64  `json:"sender_id"`
	Title     string `json:"title"`
	IsRead    bool   `json:"is_read"`
	CreatedAt string `json:"created_at"`
}

// MessageForms 消息信息列表
type MessageForms struct {
	Total    int64         `json:"total"`
	Messages []MessageForm `json:"messages"`
}

// MessageGetByIDParams 获取消息详情参数
type MessageGetByIDParams struct {
	MessageID int64 `json:"-"`
}

// MessageDetail 消息信息
type MessageDetail struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	SenderID    int64  `json:"sender_id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	RelatedID   int64  `json:"related_id"`
	RelatedType string `json:"related_type"`
	IsRead      bool   `json:"is_read"`
	CreatedAt   string `json:"created_at"`
}

// MessageReadedParams 标记消息为已读
type MessageReadedParams struct {
	UserID    int64 `json:"user_id"`
	MessageID int64 `json:"message_id"`
}

// MessageUnreadCount 未读信息总数
type MessageUnreadCount struct {
	Count int64 `json:"count"`
}
