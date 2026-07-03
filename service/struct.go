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
	NetID     string `json:"netid"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
	Level     int    `json:"level"`
}

//更新昵称
type UserUpdateParams struct {
	NetID    string `json:"-"`
	Nickname string `json:"nickname" binding:"optional,max=20"`
}

//更新头像
type UserUploadAvatar struct {
	NetID      string                `form:"-"`
	AvatarFile *multipart.FileHeader `form:"avatar" binding:"required"`
}
