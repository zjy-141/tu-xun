package controller

type Controller struct {
	Test
	User
	Activity
	Photo
	Attempt
	Admin
	Comment
	Message
	Like
	Score
	Good
	AdminGood
	AdminExchange
	AdminActivity
	Exchange
	Feedback
	Announcement
	ContentBlock
}

// New 创建并返回聚合所有子控制器的 Controller 实例
func New() *Controller {
	Controller := &Controller{}
	return Controller
}
