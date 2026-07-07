package controller

type Controller struct {
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
	Exchange
}

// New 创建并返回聚合所有子控制器的 Controller 实例
func New() *Controller {
	Controller := &Controller{}
	return Controller
}
