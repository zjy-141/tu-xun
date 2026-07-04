package service

import "mime/multipart"

// ==================== 公共 ====================

type Redirect struct {
	Redirect_url string `json:"redirect_url" form:"redirect_url" uri:"redirect_url"`
}

type Guid struct {
	Guid string `json:"guid" form:"guid" uri:"guid" binding:"required"`
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

//更新昵称
type UserUpdateParams struct {
	ID       int64  `json:"-"`
	Nickname string `json:"nickname" binding:"optional,max=20"`
}

//更新头像
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

// CreatePhotoParams 上传图片参数
type PhotoCreateParams struct {
	ID          int64                 `json:"-"`
	ActivityID  int64                 `json:"activity_id" binding:"required"`
	Title       string                `json:"title" binding:"required"`
	Description string                `json:"description"`
	ImageFile   *multipart.FileHeader `json:"image" binding:"required"`
	Longitude   float64               `json:"longitude" binding:"required"`
	Latitude    float64               `json:"latitude" binding:"required"`
}
