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
type ActivityBeief struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// ==================== Test ====================

// 测试登录
type TsetLoginParams struct {
	NetID    string `form:"netid"`
	Username string `form:"username"`
	Password string `form:"password" binding:"required"`
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
	UserID      int64                 `form:"-"`
	ActivityID  int64                 `form:"activity_id" binding:"required"`
	Title       string                `form:"title" binding:"required"`
	Description string                `form:"description" binding:"omitempty"`
	ImageFile   *multipart.FileHeader `form:"image_file" binding:"required"`
	Longitude   float64               `form:"longitude" binding:"required"`
	Latitude    float64               `form:"latitude" binding:"required"`
	CoordType   string                `form:"coord_type" binding:"required"`
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
	CreatedAt  string    `json:"created_at"`
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
	ID            int64         `json:"id"`
	Author        UserBrief     `json:"author"`
	Activity      ActivityBeief `json:"activity"`
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	ImageURL      string        `json:"image_url"`
	Solved        bool          `json:"solved"`
	AttemptsCount int           `json:"attempts_count"`
	LikesCount    int           `json:"likes_count"`
	CreatedAt     string        `json:"created_at"`
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

// PhotoAttemptsListParams 获取图片答题记录列表参数
type PhotoAttemptsUserListParams struct {
	common.PagerForm
	UserID  int64  `form:"-"`
	PhotoID int64  `form:"-"`
	SortBy  string `form:"sort_by" binding:"omitempty,oneof=created_at likes_count"`
	Status  *bool  `form:"status" binding:"omitempty,oneof=pending unsolved solved"`
}

// PhotosListUserParams 获取该用户投稿的图片列表
type PhotosListUserParams struct {
	common.PagerForm
	UserID     int64  `form:"-"`
	ActivityID int64  `form:"activity_id" binding:"required"`
	Solved     *bool  `form:"solved" binding:"omitempty,oneof=pending approved rejected"`
	SortBy     string `form:"sort_by" binding:"omitempty,oneof=created_at likes_count attempts_count"`
}

// UserPhotoForm 图片信息
type UserPhotoForm struct {
	ID int64 `json:"id"`
	// Author       UserBrief `json:"author"`
	Activity     ActivityBeief `json:"activity"`
	Title        string        `json:"title"`
	Description  string        `json:"description"`
	ThumbURL     string        `json:"thumb_url"`
	Solved       bool          `json:"solved"`
	LikesCount   int           `json:"likes_count"`
	CreatedAt    string        `json:"created_at"`
	Status       string        `json:"status"`
	RejectReason string        `json:"reject_reason"`
}

// UserPhotoForms 图片信息列表

type UserPhotoForms struct {
	Total  int64           `json:"total"`
	Photos []UserPhotoForm `json:"photos"`
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
	Activity      ActivityBeief `json:"activity"`
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	ImageURL      string        `json:"image_url"`
	Longitude     float64       `json:"longitude"`
	Latitude      float64       `json:"latitude"`
	Solved        bool          `json:"solved"`
	LikesCount    int           `json:"likes_count"`
	AttemptsCount int           `json:"attempts_count"`
	CreatedAt     string        `json:"created_at"`
	Status        string        `json:"status"`
	RejectReason  string        `json:"reject_reason"`
}

// ==================== Attempt ====================

// AttemptCreateParams 提交答题参数
type AttemptCreateParams struct {
	UserID      int64                 `form:"-"`
	PhotoID     int64                 `form:"-"`
	CommentText string                `form:"comment_text" binding:"omitempty,max=500"`
	ImageFile   *multipart.FileHeader `form:"image_file" binding:"required"`
	Longitude   float64               `form:"longitude" binding:"required"`
	Latitude    float64               `form:"latitude" binding:"required"`
	CoordType   string                `form:"coord_type" binding:"required"`
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

// AttempstListUserParams 获取该用户投稿的图片列表
type AttemptsListUserParams struct {
	common.PagerForm
	UserID int64  `form:"-"`
	Status *bool  `form:"status" binding:"omitempty,oneof=pending unsolved solved"`
	SortBy string `form:"sort_by" binding:"omitempty,oneof=created_at likes_count"`
}

// UserAttemptForm 答题信息
type UserAttemptForm struct {
	ID int64 `json:"id"`
	// Author       UserBrief `json:"author"`
	PhotoID      int64   `json:"photo_id"`
	CommentText  string  `json:"comment_text"`
	ImageURL     string  `json:"image_url"`
	Longitude    float64 `json:"longitude"`
	Latitude     float64 `json:"latitude"`
	LikesCount   int     `json:"likes_count"`
	CreatedAt    string  `json:"created_at"`
	Status       string  `json:"status"`
	RejectReason string  `json:"reject_reason"`
}

// UserAttemptForms 答题信息列表
type UserAttemptForms struct {
	Total    int64             `json:"total"`
	Attempts []UserAttemptForm `json:"attempts"`
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

type LikeCount struct {
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
	Content   string `json:"content"`
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

// MessageNoticeParams 获取公告
type MessageNoticeParams struct {
	common.PagerForm
	ActivityID int64 `form:"activity_id"`
}

// NoticeForm 公告消息信息
type NoticeForm struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	ActivityID int64  `json:"activity_id"`
	CreatedAt  string `json:"created_at"`
}

// NoticeForms 公告消息信息列表
type NoticeForms struct {
	Total   int64        `json:"total"`
	Notices []NoticeForm `json:"message_notices"`
}

// ==================== Feedback ====================

// FeedbackCreateParams 发送反馈
type FeedbackCreateParams struct {
	UserID     int64                 `form:"-"`
	Title      string                `form:"title" binding:"required,max=100"`
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
	Type   int    `form:"type" binding:"omitempty,oneof=1 2 3 4"` // 1内容 2玩法 3技术 4其他
	Status string `form:"status" binding:"omitempty,oneof=pending resolved"`
}

// FeedbackForm 反馈列表项
type FeedbackForm struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Type      int    `json:"type"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// FeedbackForms 反馈列表
type FeedbackForms struct {
	Total     int64          `json:"total"`
	Feedbacks []FeedbackForm `json:"feedbacks"`
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
	CreatedAt string              `json:"created_at"`
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

// AdminPendingPhotoParams 审核图片
type AdminPendingPhotoParams struct {
	common.PagerForm
	Status     string `form:"status" binding:"omitempty,oneof=pending approved rejected"`
	AdminLevel int    //审核员等级
}

// AdminPendingPhotoForm 待审核图片项
type AdminPendingPhotoForm struct {
	ID          int64   `json:"id"`
	UserID      int64   `json:"user_id"`
	ActivityID  int64   `json:"activity_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Longitude   float64 `json:"longitude"`
	Latitude    float64 `json:"latitude"`
	ThumbURL    string  `json:"thumb_url"`
}

type AdminPendingPhotoForms struct {
	Total         int64                   `json:"total"`
	PendingPhotos []AdminPendingPhotoForm `json:"pending_photos"`
}

// ReviewPhotoParams 审核图片参数
type AdminReviewPhotoParams struct {
	PhotoID      int64
	Action       string `json:"action" binding:"required"`
	RejectReason string `json:"reject_reason"`
	AdminLevel   int    //审核员等级
}

// AdminPendingAttemptParams 输入
type AdminPendingAttemptParams struct {
	common.PagerForm
	Status     string `form:"status" binding:"omitempty,oneof=pending approved rejected"`
	AdminLevel int    //审核员等级
}

// AdminPendingAttemptForm 待审核答题项
type AdminPendingAttemptForm struct {
	AttemptID      int64   `json:"attempt_id"`
	PhotoID        int64   `json:"photo_id"`
	PhotoTitle     string  `json:"photo_title"`
	GuassThumbURL  string  `json:"guass_image_url"` // 猜测照片
	GuassLongitude float64 `json:"guass_longitude"` // 猜测经度
	GuassLatitude  float64 `json:"guass_latitude"`  // 猜测纬度
	ThumbURL       string  `json:"thumb_url"`       // 原照片缩略图
	Longitude      float64 `json:"longitude"`       // 原照片经度（仅管理员可见）
	Latitude       float64 `json:"latitude"`        // 原照片纬度（仅管理员可见）
	Status         string  `json:"status"`          // 是否破解成功（仅管理员可见）
	SubmittedAt    string  `json:"submitted_at"`
}

// AdminPendingAttemptsResponse 待审核答题列表响应
type AdminPendingAttemptForms struct {
	Total    int64                     `json:"total"`
	Attempts []AdminPendingAttemptForm `json:"items"`
}

// AdminReviewAttemptParams 审核答题参数
type AdminReviewAttemptParams struct {
	AttemptID int64
	Solved    string `json:"solved" binding:"required,oneof=solved unsolved"` //管理员审核时是否标记图片为已破解（仅审核通过时有效）
	// Action       string `json:"action" binding:"required,oneof=approve reject"`
	RejectReason string `json:"reject_reason"  binding:"omitempty"`
	AdminLevel   int    //审核员等级
}

// AdminPendingComments:
type AdminPendingCommentParams struct {
	common.PagerForm
	Status     string `form:"status" binding:"omitempty,oneof=pending approved rejected"`
	AdminLevel int    //审核员等级
}

// AdminPendingCommentItem 待审核评论项
type AdminPendingCommentForm struct {
	CommentID  int64     `json:"comment_id"`
	PhotoID    int64     `json:"photo_id"`
	PhotoTitle string    `json:"photo_title"`
	User       UserBrief `json:"user"`
	Comment    string    `json:"comment"`
	CreatedAt  string    `json:"created_at"`
}

// AdminPendingCommentsResponse 待审核评论列表响应
type AdminPendingCommentForms struct {
	Total int64                     `json:"total"`
	Items []AdminPendingCommentForm `json:"items"`
}

// AdminReviewComment:
// AdminReviewCommentParams 审核评论参数
type AdminReviewCommentParams struct {
	CommentID    int64
	Action       string `json:"action" binding:"required"`
	RejectReason string `json:"reject_reason"`
}

// UpdateAdminLevelParams 高级管理员调整管理员等级参数
type AdminUpdateLevelParams struct {
	UserID        int64
	TargetLevel   int   `json:"target_level" binding:"required,min=0"`
	OperatorID    int64 `json:"-"` // 操作者 ID，由 controller 注入
	OperatorLevel int   `json:"-"` // 操作者等级，由 controller 注入
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
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ThumbURL    string `json:"thumb_url"`
	NeedScore   int    `json:"need_score"`
	Stock       int    `json:"stock"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

// GoodForms 奖品信息列表
type AdminGoodForms struct {
	Total int64           `json:"total"`
	Goods []AdminGoodForm `json:"goods"`
}

// GoodGetByIDParams 获取奖品详情参数
type AdminGoodGetByIDParams struct {
	GoodID int64 `json:"-"`
}

type AdminGoodDetail struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	NeedScore   int    `json:"need_score"`
	Stock       int    `json:"stock"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

// GoodCreateParams 创建商品参数
type GoodCreateParams struct {
	Name        string                `form:"name" binding:"required,max=50"`
	Description string                `form:"description" binding:"omitempty,max=500"`
	NeedScore   int                   `form:"need_score" binding:"required,min=0"`
	Stock       int                   `form:"stock" binding:"required,min=0"`
	ImageFile   *multipart.FileHeader `form:"image" binding:"required"`
	Status      string                `form:"status" binding:"omitempty,oneof=inStore outStore"`
}

// GoodUpdateParams 更新商品参数
type GoodUpdateParams struct {
	GoodID      int64                 `form:"-"`
	Name        string                `form:"name" binding:"omitempty,max=50"`
	Description string                `form:"description" binding:"omitempty,max=500"`
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
	Status string `form:"status" binding:"omitempty,oneof=pending verified cancelled"`
}
type AdminExchangeVerifyParams struct {
	ExchangeID int64  `json:"exchange_id" binding:"required"`
	Action     string `json:"action" binding:"required,oneof=verify cancel"`
}

// ExchangeForm 兑换信息
type AdminExchangeForm struct {
	ID         int64     `json:"id"`
	User       UserBrief `json:"user"`
	Good       GoodForm  `json:"good"`
	Quantity   int       `json:"quantity"`
	ScoreCost  int       `json:"score_cost"`
	Status     string    `json:"status"`
	ExchangeAt string    `json:"exchange_at"`
	CreatedAt  string    `json:"created_at"`
}

// ExchangeForms 兑换信息列表
type AdminExchangeForms struct {
	Total          int64               `json:"total"`
	AdminExchanges []AdminExchangeForm `json:"exchanges"`
}

// ==================== AdminActivity ====================

// RewardTierInput 奖励阶梯入参
type RewardTierInput struct {
	Batch         int `json:"batch" binding:"required,min=1"`
	RankLimit     int `json:"rank_limit" binding:"required,min=1"`
	AttemptPoints int `json:"attempt_points" binding:"required,min=0"`
}

// AdminActivityCreate 创建活动参数
type AdminActivityCreate struct {
	Title       string                `form:"title" binding:"required,max=255"`
	CoverFile   *multipart.FileHeader `form:"cover_file" binding:"omitempty"`
	Description string                `form:"description" binding:"required"`
	StartTime   string                `form:"start_time" binding:"required"`
	EndTime     string                `form:"end_time" binding:"required"`
	PhotoPoints *int                  `form:"photo_points" binding:"required"`
	RewardTiers string                `form:"reward_tiers" binding:"omitempty"` // 改为 string
}

// AdminActivityUpdate 更新活动参数
type AdminActivityUpdate struct {
	ActivityID  int64                 `form:"activity_id" binding:"required"`
	Title       string                `form:"title" binding:"omitempty,max=255"`
	CoverFile   *multipart.FileHeader `form:"cover_file" binding:"omitempty"`
	Description string                `form:"description" binding:"omitempty"`
	StartTime   string                `form:"start_time" binding:"omitempty"`
	EndTime     string                `form:"end_time" binding:"omitempty"`
	IsActive    bool                  `form:"is_active" binding:"omitempty"`
	PhotoPoints *int                  `form:"photo_points" binding:"omitempty,min=0"`
	RewardTiers string                `form:"reward_tiers" binding:"omitempty"` // 改为 string
}

// AdminActivityNotice 活动公告参数
type AdminActivityNotice struct {
	ActivityID int64  `json:"activity_id" binding:"required"`
	Title      string `json:"title" binding:"required,max=128"`
	Content    string `json:"content" binding:"required"`
}
